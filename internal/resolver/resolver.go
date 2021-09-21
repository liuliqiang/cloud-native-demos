package resolver

type Resolver interface {
	Resolve(addr string) (endpoints []string, err error)
}
