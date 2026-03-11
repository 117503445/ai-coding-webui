package main

var cli struct {
	Build  cmdBuild  `cmd:"" help:"Build binaries"`
	Format cmdFormat `cmd:"" help:"Format and tidy project"`
}

type cmdBuild struct{}

func (c *cmdBuild) Run() error {
	build()
	return nil
}

type cmdFormat struct{}

func (c *cmdFormat) Run() error {
	format()
	return nil
}
