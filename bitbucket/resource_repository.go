package bitbucket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

// CloneURL is the internal struct we use to represent urls
type CloneURL struct {
	Href string `json:"href,omitempty"`
	Name string `json:"name,omitempty"`
}

// PipelinesEnabled is the struct we send to turn on or turn off pipelines for a repository
type PipelinesEnabled struct {
	Enabled bool `json:"enabled"`
}

// DevelopmentBranch is the struct we send to configure development branch for a repository
type DevelopmentBranch struct {
	IsValid       bool   `json:"is_valid,omitempty"`
	Name          string `json:"name,omitempty"`
	UseMainbranch bool   `json:"use_mainbranch"`
}

// ProductionBranch is the struct we send to configure production branch for a repository
type ProductionBranch struct {
	IsValid       bool   `json:"is_valid,omitempty"`
	Enabled       bool   `json:"enabled,omitempty"`
	Name          string `json:"name,omitempty"`
	UseMainbranch bool   `json:"use_mainbranch,omitempty"`
}

// BranchType is the struct we send to configure branch types for a repository
type BranchType struct {
	Kind    string `json:"kind,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
	Prefix  string `json:"prefix,omitempty"`
}

// BranchingModelSettings is the struct we send to configure branching model for a repository
type BranchingModelSettings struct {
	Development DevelopmentBranch `json:"development,omitempty"`
	Production  ProductionBranch  `json:"production,omitempty"`
	BranchTypes []BranchType      `json:"branch_types,omitempty"`
}

// Repository is the struct we need to send off to the Bitbucket API to create a repository
type Repository struct {
	SCM         string `json:"scm,omitempty"`
	HasWiki     bool   `json:"has_wiki,omitempty"`
	HasIssues   bool   `json:"has_issues,omitempty"`
	Website     string `json:"website,omitempty"`
	IsPrivate   bool   `json:"is_private,omitempty"`
	ForkPolicy  string `json:"fork_policy,omitempty"`
	Language    string `json:"language,omitempty"`
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
	Slug        string `json:"slug,omitempty"`
	UUID        string `json:"uuid,omitempty"`
	Project     struct {
		Key string `json:"key,omitempty"`
	} `json:"project,omitempty"`
	Links struct {
		Clone []CloneURL `json:"clone,omitempty"`
	} `json:"links,omitempty"`
}

func resourceRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceRepositoryCreate,
		Update: resourceRepositoryUpdate,
		Read:   resourceRepositoryRead,
		Delete: resourceRepositoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"scm": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "git",
			},
			"has_wiki": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"has_issues": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"website": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"clone_ssh": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"clone_https": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_private": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"pipelines_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"fork_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "allow_forks",
			},
			"language": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"branching_model_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"development": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"use_mainbranch": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
								},
							},
						},
						"branch_types": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 4,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"kind": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											"release",
											"hotfix",
											"feature",
											"bugfix",
										},
											false),
									},
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"production": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"use_mainbranch": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func newRepositoryFromResource(d *schema.ResourceData) *Repository {
	repo := &Repository{
		Name:        d.Get("name").(string),
		Slug:        d.Get("slug").(string),
		Language:    d.Get("language").(string),
		IsPrivate:   d.Get("is_private").(bool),
		Description: d.Get("description").(string),
		ForkPolicy:  d.Get("fork_policy").(string),
		HasWiki:     d.Get("has_wiki").(bool),
		HasIssues:   d.Get("has_issues").(bool),
		SCM:         d.Get("scm").(string),
		Website:     d.Get("website").(string),
	}

	repo.Project.Key = d.Get("project_key").(string)
	return repo
}

func resourceRepositoryUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)
	repository := newRepositoryFromResource(d)

	var jsonbuffer []byte

	jsonpayload := bytes.NewBuffer(jsonbuffer)
	enc := json.NewEncoder(jsonpayload)
	enc.Encode(repository)

	var repoSlug string
	repoSlug = d.Get("slug").(string)
	if repoSlug == "" {
		repoSlug = d.Get("name").(string)
	}

	_, err := client.Put(fmt.Sprintf("2.0/repositories/%s/%s",
		d.Get("owner").(string),
		repoSlug,
	), jsonpayload)

	if err != nil {
		return err
	}

	var pipelinesEnabled bool
	pipelinesEnabled = d.Get("pipelines_enabled").(bool)
	pipelinesConfig := &PipelinesEnabled{Enabled: pipelinesEnabled}

	bytedata, err := json.Marshal(pipelinesConfig)

	if err != nil {
		return err
	}

	_, err = client.Put(fmt.Sprintf("2.0/repositories/%s/%s/pipelines_config",
		d.Get("owner").(string),
		repoSlug), bytes.NewBuffer(bytedata))

	if err != nil {
		return err
	}

	branchingModelSettings := d.Get("branching_model_settings").([]interface{})
	if len(branchingModelSettings) == 0 || branchingModelSettings[0] == nil {
	} else {
		loc := branchingModelSettings[0].(map[string]interface{})

		outBranchingModelSettings := &BranchingModelSettings{}

		developmentBranch := loc["development"].([]interface{})
		if len(developmentBranch) == 0 || developmentBranch[0] == nil {
		} else {
			loc2 := developmentBranch[0].(map[string]interface{})
			outDevelopmentBranch := &DevelopmentBranch{}
			if v, ok := loc2["name"]; ok {
				outDevelopmentBranch.Name = v.(string)
			}
			if v, ok := loc2["use_mainbranch"]; ok {
				outDevelopmentBranch.UseMainbranch = v.(bool)
			}

			outBranchingModelSettings.Development = *outDevelopmentBranch
		}

		productionBranch := loc["production"].([]interface{})
		if len(productionBranch) == 0 || productionBranch[0] == nil {
		} else {
			loc2 := productionBranch[0].(map[string]interface{})
			outProductionBranch := &ProductionBranch{}
			if v, ok := loc2["name"]; ok {
				outProductionBranch.Name = v.(string)
			}
			if v, ok := loc2["use_mainbranch"]; ok {
				outProductionBranch.UseMainbranch = v.(bool)
			}
			if v, ok := loc2["enabled"]; ok {
				outProductionBranch.Enabled = v.(bool)
			}

			outBranchingModelSettings.Production = *outProductionBranch
		}

		branchTypes := loc["branch_types"].([]interface{})
		if len(branchTypes) == 0 || branchTypes[0] == nil {
		} else {
			outBranchTypes := make([]BranchType, 0, len(branchTypes))

			for _, item := range branchTypes {
				loc2 := item.(map[string]interface{})
				outBranchType := &BranchType{}
				if v, ok := loc2["kind"]; ok {
					outBranchType.Kind = v.(string)
				}
				if v, ok := loc2["enabled"]; ok {
					outBranchType.Enabled = v.(bool)
				}
				if v, ok := loc2["prefix"]; ok {
					outBranchType.Prefix = v.(string)
				}

				outBranchTypes = append(outBranchTypes, *outBranchType)
			}

			outBranchingModelSettings.BranchTypes = outBranchTypes
		}

		bytedata, err = json.Marshal(outBranchingModelSettings)

		if err != nil {
			return err
		}

		_, err = client.Put(fmt.Sprintf("2.0/repositories/%s/%s/branching-model/settings",
			d.Get("owner").(string),
			repoSlug), bytes.NewBuffer(bytedata))

		if err != nil {
			return err
		}
	}

	return resourceRepositoryRead(d, m)
}

func resourceRepositoryCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)
	repo := newRepositoryFromResource(d)

	bytedata, err := json.Marshal(repo)

	if err != nil {
		return err
	}

	var repoSlug string
	repoSlug = d.Get("slug").(string)
	if repoSlug == "" {
		repoSlug = d.Get("name").(string)
	}

	_, err = client.Post(fmt.Sprintf("2.0/repositories/%s/%s",
		d.Get("owner").(string),
		repoSlug,
	), bytes.NewBuffer(bytedata))

	if err != nil {
		return err
	}
	d.SetId(string(fmt.Sprintf("%s/%s", d.Get("owner").(string), repoSlug)))

	var pipelinesEnabled bool
	pipelinesEnabled = d.Get("pipelines_enabled").(bool)
	pipelinesConfig := &PipelinesEnabled{Enabled: pipelinesEnabled}

	bytedata, err = json.Marshal(pipelinesConfig)

	if err != nil {
		return err
	}

	_, err = client.Put(fmt.Sprintf("2.0/repositories/%s/%s/pipelines_config",
		d.Get("owner").(string),
		repoSlug), bytes.NewBuffer(bytedata))

	if err != nil {
		return err
	}

	branchingModelSettings := d.Get("branching_model_settings").([]interface{})
	if len(branchingModelSettings) == 0 || branchingModelSettings[0] == nil {
	} else {
		loc := branchingModelSettings[0].(map[string]interface{})

		outBranchingModelSettings := &BranchingModelSettings{}

		developmentBranch := loc["development"].([]interface{})
		if len(developmentBranch) == 0 || developmentBranch[0] == nil {
		} else {
			loc2 := developmentBranch[0].(map[string]interface{})
			outDevelopmentBranch := &DevelopmentBranch{}
			if v, ok := loc2["name"]; ok {
				outDevelopmentBranch.Name = v.(string)
			}
			if v, ok := loc2["use_mainbranch"]; ok {
				outDevelopmentBranch.UseMainbranch = v.(bool)
			}
			outBranchingModelSettings.Development = *outDevelopmentBranch
		}

		productionBranch := loc["production"].([]interface{})
		if len(productionBranch) == 0 || productionBranch[0] == nil {
		} else {
			loc2 := productionBranch[0].(map[string]interface{})
			outProductionBranch := &ProductionBranch{}
			if v, ok := loc2["name"]; ok {
				outProductionBranch.Name = v.(string)
			}
			if v, ok := loc2["use_mainbranch"]; ok {
				outProductionBranch.UseMainbranch = v.(bool)
			}
			if v, ok := loc2["enabled"]; ok {
				outProductionBranch.Enabled = v.(bool)
			}

			outBranchingModelSettings.Production = *outProductionBranch
		}

		bytedata, err = json.Marshal(outBranchingModelSettings)

		if err != nil {
			return err
		}

		_, err = client.Put(fmt.Sprintf("2.0/repositories/%s/%s/branching-model/settings",
			d.Get("owner").(string),
			repoSlug), bytes.NewBuffer(bytedata))

		if err != nil {
			return err
		}
	}

	return resourceRepositoryRead(d, m)
}

func resourceRepositoryRead(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id != "" {
		idparts := strings.Split(id, "/")
		if len(idparts) == 2 {
			d.Set("owner", idparts[0])
			d.Set("slug", idparts[1])
		} else {
			return fmt.Errorf("Incorrect ID format, should match `owner/slug`")
		}
	}

	var repoSlug string
	repoSlug = d.Get("slug").(string)
	if repoSlug == "" {
		repoSlug = d.Get("name").(string)
	}

	client := m.(*Client)
	repoReq, _ := client.Get(fmt.Sprintf("2.0/repositories/%s/%s",
		d.Get("owner").(string),
		repoSlug,
	))

	if repoReq.StatusCode == 200 {

		var repo Repository

		body, readerr := ioutil.ReadAll(repoReq.Body)
		if readerr != nil {
			return readerr
		}

		decodeerr := json.Unmarshal(body, &repo)
		if decodeerr != nil {
			return decodeerr
		}

		d.Set("scm", repo.SCM)
		d.Set("is_private", repo.IsPrivate)
		d.Set("has_wiki", repo.HasWiki)
		d.Set("has_issues", repo.HasIssues)
		d.Set("name", repo.Name)
		if repo.Slug != "" && repo.Name != repo.Slug {
			d.Set("slug", repo.Slug)
		}
		d.Set("language", repo.Language)
		d.Set("fork_policy", repo.ForkPolicy)
		d.Set("website", repo.Website)
		d.Set("description", repo.Description)
		d.Set("project_key", repo.Project.Key)

		for _, cloneURL := range repo.Links.Clone {
			if cloneURL.Name == "https" {
				d.Set("clone_https", cloneURL.Href)
			} else {
				d.Set("clone_ssh", cloneURL.Href)
			}
		}
		pipelinesConfigReq, err := client.Get(fmt.Sprintf("2.0/repositories/%s/%s/pipelines_config",
			d.Get("owner").(string),
			repoSlug))

		if err != nil {
			return err
		}

		if pipelinesConfigReq.StatusCode == 200 {
			var pipelinesConfig PipelinesEnabled

			body, readerr := ioutil.ReadAll(pipelinesConfigReq.Body)
			if readerr != nil {
				return readerr
			}

			decodeerr := json.Unmarshal(body, &pipelinesConfig)
			if decodeerr != nil {
				return decodeerr
			}

			d.Set("pipelines_enabled", pipelinesConfig.Enabled)
		}

		branchingModelSettingsReq, err := client.Get(fmt.Sprintf("2.0/repositories/%s/%s/branching-model/settings",
			d.Get("owner").(string),
			repoSlug))

		if err != nil {
			return err
		}

		if branchingModelSettingsReq.StatusCode == 200 {
			var branchingModelSettings BranchingModelSettings

			body, readerr := ioutil.ReadAll(branchingModelSettingsReq.Body)
			if readerr != nil {
				return readerr
			}

			decodeerr := json.Unmarshal(body, &branchingModelSettings)
			if decodeerr != nil {
				return decodeerr
			}

			d.Set("branching_model_settings", flattenBranchingModelSettings(&branchingModelSettings))
		}
	}

	return nil
}

func resourceRepositoryDelete(d *schema.ResourceData, m interface{}) error {

	var repoSlug string
	repoSlug = d.Get("slug").(string)
	if repoSlug == "" {
		repoSlug = d.Get("name").(string)
	}

	client := m.(*Client)
	_, err := client.Delete(fmt.Sprintf("2.0/repositories/%s/%s",
		d.Get("owner").(string),
		repoSlug,
	))

	return err
}

func flattenBranchingModelSettings(in *BranchingModelSettings) []interface{} {
	m := make(map[string]interface{})
	m["branch_types"] = flattenBranchingModelSettingsBranchTypes(in.BranchTypes)
	m["development"] = flattenBranchingModelSettingsDevelopmentBranch(&in.Development)
	m["production"] = flattenBranchingModelSettingsProductionBranch(&in.Production)

	return []interface{}{m}
}

func flattenBranchingModelSettingsBranchTypes(in []BranchType) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(in))

	for _, item := range in {
		out = append(out, flattenBranchingModelSettingsBranchType(&item))
	}

	return out
}

func flattenBranchingModelSettingsBranchType(in *BranchType) map[string]interface{} {
	out := make(map[string]interface{})
	out["kind"] = in.Kind
	out["enabled"] = in.Enabled
	out["prefix"] = in.Prefix

	return out
}

func flattenBranchingModelSettingsDevelopmentBranch(in *DevelopmentBranch) []interface{} {
	out := make(map[string]interface{})
	out["name"] = in.Name
	out["use_mainbranch"] = in.UseMainbranch

	return []interface{}{out}
}

func flattenBranchingModelSettingsProductionBranch(in *ProductionBranch) []interface{} {
	out := make(map[string]interface{})
	out["enabled"] = in.Enabled
	out["name"] = in.Name
	out["use_mainbranch"] = in.UseMainbranch

	return []interface{}{out}
}
