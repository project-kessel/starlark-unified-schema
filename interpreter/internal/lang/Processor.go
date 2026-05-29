package lang

import (
	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"go.starlark.net/starlark"
)

type Processor struct {
	thread *starlark.Thread
	loader *Loader
}

func NewProcessor(loader *Loader) *Processor {
	return &Processor{
		loader: loader,
		thread: &starlark.Thread{Name: "processor thread", Load: loader.Load},
	}
}

func (p *Processor) ProcessModule(name string, visitor output.Visitor) error {
	globals, err := p.loader.Load(p.thread, name)
	if err != nil {
		return err
	}

	// Evaluate globals and call appropriate visitor methods
	_ = globals

	return nil
}

func (p *Processor) ProcessAllModules(visitor output.Visitor) error {
	names, err := p.loader.GetAllModuleNames()
	if err != nil {
		return err
	}

	for _, name := range names {
		err := p.ProcessModule(name, visitor)
		if err != nil {
			return err
		}
	}

	return nil
}
