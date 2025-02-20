package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	elastic7 "github.com/olivere/elastic/v7"
)

func resourceOpensearchSnapshotRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpensearchSnapshotRepositoryCreate,
		Read:   resourceOpensearchSnapshotRepositoryRead,
		Update: resourceOpensearchSnapshotRepositoryUpdate,
		Delete: resourceOpensearchSnapshotRepositoryDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"settings": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceOpensearchSnapshotRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	err := resourceOpensearchSnapshotRepositoryUpdate(d, meta)
	if err != nil {
		return err
	}
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceOpensearchSnapshotRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var repositoryType string
	var settings map[string]interface{}
	var err error
	client, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	repositoryType, settings, err = elastic7SnapshotGetRepository(client, id)
	if err != nil {
		return err
	}

	ds := &resourceDataSetter{d: d}
	ds.set("name", id)
	ds.set("type", repositoryType)
	ds.set("settings", settings)
	return ds.err
}

func elastic7SnapshotGetRepository(client *elastic7.Client, id string) (string, map[string]interface{}, error) {
	repos, err := client.SnapshotGetRepository(id).Do(context.TODO())
	if err != nil {
		return "", make(map[string]interface{}), err
	}

	return repos[id].Type, repos[id].Settings, nil
}

func resourceOpensearchSnapshotRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	repositoryType := d.Get("type").(string)
	name := d.Get("name").(string)

	var settings map[string]interface{}

	if v, ok := d.GetOk("settings"); ok {
		settings = v.(map[string]interface{})
	}

	var err error
	client, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	err = elastic7SnapshotCreateRepository(client, name, repositoryType, settings)

	return err
}

func elastic7SnapshotCreateRepository(client *elastic7.Client, name string, repositoryType string, settings map[string]interface{}) error {
	repo := elastic7.SnapshotRepositoryMetaData{
		Type:     repositoryType,
		Settings: settings,
	}

	_, err := client.SnapshotCreateRepository(name).BodyJson(&repo).Do(context.TODO())
	return err
}

func resourceOpensearchSnapshotRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()

	var err error
	client, err := getClient(meta.(*ProviderConf))
	if err != nil {
		return err
	}
	err = elastic7SnapshotDeleteRepository(client, id)

	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func elastic7SnapshotDeleteRepository(client *elastic7.Client, id string) error {
	_, err := client.SnapshotDeleteRepository(id).Do(context.TODO())
	return err
}
