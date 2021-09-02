# tf-tags
tf-tags helps to use structs to set and get values from terraform schema. This will helps developers to use same struct
to parse json request and use the same struct for terraform as well.

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
        "calc": {
            Type:        schema.TypeInt,
            Computed:    true,
            Default:     "Default",
        }
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

Nested structs are supported for nested terraform resources.

_**Note:** Please note that, you need to provide `tf` tags for every fields to parse, any fields without `tf` tags will be omitted_

You can use `Set` function to set computed fields of a resource. To denote computed field use `computed` in the tags.

example:
```
    type Model struct {
        Id         string `tf:"id"`
        Site       string `tf:"site"`
        Name       string `tf:"name"`
        Calculated int `tf:"calc,computed"
    }

    err:= tftags.Set(d,Model{Calculated: 5})
```

_**Note:** You need to set computed on top level field of the struct, otherwise it won't work_


# License
https://github.com/tshihad/tftags/blob/main/LICENSE
