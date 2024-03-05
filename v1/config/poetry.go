package config

import (
	"fmt"
	"net/mail"

	"github.com/BurntSushi/toml"
)

type Poetry struct {
	Authors      []PoetryAuthor              `toml:"authors"`
	Name         string                      `toml:"name"`
	Description  string                      `toml:"description"`
	Dependencies map[string]PoetryDependency `toml:"dependencies"`
}

func (p *Poetry) GetAuthors() []Author {
	var authors []Author
	for _, a := range p.Authors {
		authors = append(authors, a.ToAuthor())
	}
	return authors
}

func (p *Poetry) PythonRequires() string {
	if v, ok := p.Dependencies["python"]; ok {
		return v.version
	}
	return ""
}

type PoetryAuthor struct {
	Name  string `toml:"name"`
	Email string `toml:"email"`
}

func (p *PoetryAuthor) ToAuthor() Author {
	return Author{
		Name:  p.Name,
		Email: p.Email,
	}
}

func (p *PoetryAuthor) UnmarshalTOML(value interface{}) error {
	text, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", value)
	}
	addr, err := mail.ParseAddress(text)
	p.Email = addr.Address
	p.Name = addr.Name
	return err
}

type PoetryDependency struct {
	version string
}

func (p *PoetryDependency) UnmarshalTOML(value interface{}) error {
	text, ok := value.(string)
	if ok {
		p.version = text
		return nil
	}
	mapping, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected string or map, got %T", value)
	}
	if v, ok := mapping["version"]; ok {
		p.version = v.(string)
	} else {
		return fmt.Errorf("version field is required")
	}
	return nil
}

var (
	_ toml.Unmarshaler = (*PoetryAuthor)(nil)
	_ toml.Unmarshaler = (*PoetryDependency)(nil)
)
