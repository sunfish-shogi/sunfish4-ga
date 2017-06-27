package ga

type Param struct {
	Name            string
	FirstEliteValue int32
	MinimumValue    int32
	MaximumValue    int32
}

type Config struct {
	Params             []Param
	NumberOfIndividual int
}
