package content

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type portfolioFile struct {
	Projects []Project `yaml:"projects"`
}

// LoadPortfolio reads and parses the portfolio data file at path.
func LoadPortfolio(path string) ([]Project, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading portfolio file %q: %w", path, err)
	}

	var pf portfolioFile
	if err := yaml.Unmarshal(raw, &pf); err != nil {
		return nil, fmt.Errorf("parsing portfolio file %q: %w", path, err)
	}
	return pf.Projects, nil
}
