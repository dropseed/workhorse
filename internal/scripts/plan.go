package scripts

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dropseed/workhorse/internal/config"
	"github.com/dropseed/workhorse/internal/meta"
	"github.com/dropseed/workhorse/internal/utils"
	"github.com/mitchellh/mapstructure"
)

func GetPlansDir() string {
	return path.Join(meta.AppName, "plans")
}

type Plan struct {
	// version of release that ran it?
	Script  string        `json:"script"`
	Targets []string      `json:"targets"`
	Config  config.Config `json:"config"`
	id      string
}

func NewPlan(script string, config config.Config) (*Plan, error) {
	return &Plan{
		Script:  script,
		Targets: []string{},
		Config:  config,
	}, nil
}

func NewPlanFromPath(path string) (*Plan, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	return NewPlanFromReader(f)
}

func NewPlanFromReader(reader io.Reader) (*Plan, error) {
	temp := map[string]interface{}{}
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&temp); err != nil {
		return nil, err
	}

	return newPlanFromMap(temp)
}

func newPlanFromMap(m map[string]interface{}) (*Plan, error) {
	plan := &Plan{}

	mapDecoderConfig := mapstructure.DecoderConfig{
		Result:      plan,
		ErrorUnused: true,
	}
	mapDecoder, err := mapstructure.NewDecoder(&mapDecoderConfig)
	if err != nil {
		return nil, err
	}

	if err = mapDecoder.Decode(m); err != nil {
		return nil, err
	}

	return plan, nil
}

func (p *Plan) Validate() error {
	if err := p.Config.Validate(); err != nil {
		return err
	}
	return nil
}

func (p *Plan) Load() error {
	targets, err := p.Config.GetTargets()
	if err != nil {
		return err
	}

	p.Targets = targets

	return nil
}

func (p *Plan) Execute() error {
	return p.Config.ExecuteTargets(p.Targets)
}

func (p *Plan) GetSlug() string {
	return fmt.Sprintf("%s-%s", meta.AppAbbr, p.getID())
}

func (p *Plan) getID() string {
	if p.id != "" {
		// Return the saved id if already known or generated
		return p.id
	}

	plansDir := GetPlansDir()
	files, err := ioutil.ReadDir(plansDir)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(plansDir, os.ModePerm); err != nil {
			panic(err)
		}
		files, err = ioutil.ReadDir(plansDir)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}

	planNumber := 1

	for _, f := range files {
		if ext := filepath.Ext(f.Name()); ext == ".json" {
			name := utils.ExtensionlessBasename(f.Name())
			if strings.HasPrefix(name, meta.AppAbbr+"-") {
				name = name[len(meta.AppAbbr)+1:]
			}
			num, err := strconv.Atoi(name)
			if err == nil && num+1 > planNumber {
				planNumber = num + 1
			}
		}
	}

	p.id = strconv.Itoa(planNumber)
	return p.id
}

func (p *Plan) GetPath() string {
	plansDir := GetPlansDir()
	return path.Join(plansDir, fmt.Sprintf("%s.json", p.GetSlug()))
}

func (p *Plan) Save() (string, error) {
	out, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	out = append(out, "\n"...)
	path := p.GetPath()
	if err := ioutil.WriteFile(path, out, 0644); err != nil {
		panic(err)
	}
	return path, nil
}
