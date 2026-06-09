package lang

import (
	"github.com/project-kessel/starlark-unified-schema/internal/output"
	"go.starlark.net/starlark"
)

type resourceType *starlark.Dict

type Processor struct {
	thread   *starlark.Thread
	loader   *Loader
	metadata map[resourceType]meta
}

func NewProcessor(loader *Loader) *Processor {
	m := map[resourceType]meta{}

	p := &Processor{
		loader:   loader,
		thread:   &starlark.Thread{Name: "processor thread", Load: loader.Load},
		metadata: m,
	}

	loader.SetMetadata(m)

	return p
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

type meta struct {
	moduleName string
	typeName   string
}
