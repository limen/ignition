package ignitor

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/limen/fargo/utils"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

var ErrPoolExhausted = errors.New("connection pool exhausted")

// 连接接口
type Conn interface {
	Close() error
}

type Pool struct {
	ID string
	// 创建连接对象
	Dial func() (Conn, error)
	// 可存在的最大连接数
	MaxActive int
	// 最大空闲连接数
	MaxIdle         int
	IdleTimeout     time.Duration
	MaxConnLifetime time.Duration
	// 是否等待
	Wait          bool
	chInitialized uint32
	mu            sync.Mutex    // mu protects the following fields
	closed        bool          // set to true when the pool is closed.
	active        int           // the number of open connections in the pool
	ch            chan struct{} // limits open connections when p.Wait is true
	idle          idleList      // idle connections
}

// 空闲链表
// 双向链表
// front指向左侧头节点
// back指向右侧头节点
type idleList struct {
	count       int
	front, back *PooledConn
}

// 池中连接
type PooledConn struct {
	Conn Conn
	// 所属连接池
	pool *Pool
	// 最后使用时间
	latestUsedAt time.Time
	// 创建时间
	createdAt  time.Time
	next, prev *PooledConn
}

func (pc *PooledConn) GetDB() *gorm.DB {
	gc, ok := interface{}(pc.Conn).(*gorm.DB)
	if !ok {
		panic("type assertion failed: " + reflect.ValueOf(pc.Conn).Type().String())
	}

	return gc
}

func (pc *PooledConn) GetRedis() *redis.Client {
	rc, ok := interface{}(pc.Conn).(*redis.Client)
	if !ok {
		panic("type assertion failed: " + reflect.ValueOf(pc.Conn).Type().String())
	}

	return rc
}

func (p *Pool) Stat() utils.MapStr {
	return utils.MapStr{
		"ID":          p.ID,
		"idleCount":   p.idle.count,
		"maxActive":   p.MaxActive,
		"maxIdle":     p.MaxIdle,
		"activeCount": p.active,
		"id":          fmt.Sprintf("%v", &p),
	}
}

// 从池中获取连接
func (p *Pool) Get() (*PooledConn, error) {
	// Handle limit for p.Wait == true.
	if p.Wait && p.MaxActive > 0 {
		p.lazyInit()
		<-p.ch
	}

	p.mu.Lock()

	// Prune stale connections at the back of the idle list.
	if p.IdleTimeout > 0 {
		n := p.idle.count
		for i := 0; i < n && p.idle.back != nil && p.idle.back.latestUsedAt.Add(p.IdleTimeout).Before(time.Now()); i++ {
			pc := p.idle.back
			p.idle.popBack()
			p.mu.Unlock()
			pc.Conn.Close()
			p.mu.Lock()
			p.active--
		}
	}

	// Get idle connection from the front of idle list.
	for p.idle.front != nil {
		pc := p.idle.front
		p.idle.popFront()
		p.mu.Unlock()
		return pc, nil
	}

	// Check for pool closed before dialing a new connection.
	if p.closed {
		p.mu.Unlock()
		return nil, errors.New("get on closed pool")
	}

	// Handle limit for p.Wait == false.
	if !p.Wait && p.MaxActive > 0 && p.active >= p.MaxActive {
		p.mu.Unlock()
		return nil, ErrPoolExhausted
	}

	p.active++
	p.mu.Unlock()
	c, err := p.Dial()
	if err != nil {
		c = nil
		p.mu.Lock()
		p.active--
		if p.ch != nil && !p.closed {
			p.ch <- struct{}{}
		}
		p.mu.Unlock()
	}
	newPc := PooledConn{pool: p, Conn: c, createdAt: time.Now()}
	return &newPc, err
}

// 关闭连接池
func (p *Pool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.active -= p.idle.count
	pc := p.idle.front
	p.idle.count = 0
	p.idle.front, p.idle.back = nil, nil
	if p.ch != nil {
		close(p.ch)
	}
	p.mu.Unlock()
	for ; pc != nil; pc = pc.next {
		pc.Conn.Close()
	}
	return nil
}

// 关闭池外连接 = 将连接归还
func (pc *PooledConn) Close() error {
	return pc.pool.put(pc)
}

// 归还连接
func (p *Pool) put(pc *PooledConn) error {
	p.mu.Lock()
	if !p.closed {
		// 最后使用时间
		pc.latestUsedAt = time.Now()
		p.idle.pushFront(pc)
		if p.idle.count > p.MaxIdle {
			pc = p.idle.back
			p.idle.popBack()
		} else {
			pc = nil
		}
	}
	// 需要关闭多出连接
	// 可能是pc本身（pool关闭时）
	// 或idle尾部连接
	if pc != nil {
		pc.Conn.Close()
		p.active--
	}
	if p.ch != nil && !p.closed {
		p.ch <- struct{}{}
	}
	p.mu.Unlock()
	return nil
}

func (p *Pool) lazyInit() {
	// Fast path.
	if atomic.LoadUint32(&p.chInitialized) == 1 {
		return
	}
	// Slow path.
	p.mu.Lock()
	if p.chInitialized == 0 {
		p.ch = make(chan struct{}, p.MaxActive)
		if p.closed {
			close(p.ch)
		} else {
			for i := 0; i < p.MaxActive; i++ {
				p.ch <- struct{}{}
			}
		}
		atomic.StoreUint32(&p.chInitialized, 1)
	}
	p.mu.Unlock()
}

func (l *idleList) pushFront(pc *PooledConn) {
	pc.next = l.front
	pc.prev = nil
	if l.count == 0 {
		l.back = pc
	} else {
		l.front.prev = pc
	}
	l.front = pc
	l.count++
	return
}

func (l *idleList) popFront() {
	pc := l.front
	l.count--
	if l.count == 0 {
		l.front, l.back = nil, nil
	} else {
		pc.next.prev = nil
		l.front = pc.next
	}
	pc.next, pc.prev = nil, nil
}

func (l *idleList) popBack() {
	pc := l.back
	l.count--
	if l.count == 0 {
		l.front, l.back = nil, nil
	} else {
		pc.prev.next = nil
		l.back = pc.prev
	}
	pc.next, pc.prev = nil, nil
}
