package ignitor

type Getter interface {
	// Since getters may be related with each other,
	// one getter may need another one to do something for it e.g.,
	// this method need to be aware of all getters.
	// Ctx should holds the data which would be passed through all getters.
	// allGetters is type of map[string]bool for easier checking if a getter exists.
	Get(ctx interface{}, allGetters map[string]bool) interface{}
}
