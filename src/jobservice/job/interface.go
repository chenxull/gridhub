package job

//Interface defines the related injection and run entry methods
type Interface interface {
	MaxFalis() uint

	ShouldRetry() bool

	Validate(params Parameters) error
	//Run the business logic here
	// The related arguments will be injected by the workerpool.
	Run(ctx Context, params Parameters) error
}
