# tf-tags
tf-tags helps to use structs to set and get values from terraform schema

## Example
Assume you had a resource as following
```
&schema.Resource{
    Schema: map[string]*schema.Schema{
        "id": {
            Type:        schema.TypeString,
            Computed:    true,
        },
        "site": {
            Type:        schema.TypeString,
            Computed:    true,
            Optional:    true,
        },
        "name": {
            Type:        schema.TypeString,
            Optional:    true,
            Default:     "Default",
        },
    },
}
```

You can parse above schema to following struct
```
    type Model struct {
        Id   string `tf:"id"`
        Site string `tf:"site"`
        Name string `tf:"name"`
    }
    var model Model
    err := tftags.Get(d, &model)
    if err!=nil{
        return err
    }
```

## TODO
- Feature for set computed field
- Other than struct, support for map also
