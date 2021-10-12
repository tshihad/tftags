# tftags
**tftags** or terraform tags is a helper package which can use for parse given terraform state to a struct as
well as set all computed fields into state file. `tftags` library contains two functions `Get` and `Set`.
`Get` used to parse schema values into struct and `Set` used to set schema data into state file

## Where to use tftags?
* If your API request schema is similar to terraform schema.
* If your schema is quite bigger and hard to manage. Rather than storing values in maps, you can store in structs

## How to use tftags
Assume you had a resource as following
```go
&schema.Resource{
    Schema: map[string]*schema.Schema{
        "id": {
            Type:     schema.TypeInt,
            Computed: true,
        },
        "class": {
            Type:     schema.TypeString,
            Computed: true,
            Optional: true,
        },
        "subject": {
            Type:     schema.TypeList,
            Required: true,
            Elem: &schema.Resource{
                Schema: map[string]*schema.Schema{
                    "name": {
                        Type:     schema.TypeString,
                        Required: true,
                    },
                    "id": {
                        Type:     schema.TypeInt,
                        Computed: true,
                    },
                },
            },
        },
        "mark": {
            Type:     schema.TypeSet,
            Computed: true,
            Elem: &schema.Resource{
                Schema: map[string]*schema.Schema{
                    "subject": {
                        Type:     schema.TypeString,
                        Computed: true,
                    },
                    "id": {
                        Type:     schema.TypeInt,
                        Computed: true,
                    },
                },
            },
        },
        "rank": {
            Type:     schema.TypeInt,
            Computed: true,
        },
    },
}
```

To parse above schema you need to create new struct with tags or if you already had one add tf tags to it

Here we created struct for parsing. Please note you can include other tags such as `json` or `bson` etc along with `tf`.

```go
    type Subject struct {
		Name string `tf:"name"`
		// Eventhough ID is computed we don't need to mark it
		// as computed since top level field is marked as computed
		ID int `tf:"id"`
	}

	type Mark struct {
		Subject string `tf:"subject"`
		// Eventhough ID is computed we don't need to mark it
		// as computed since top level field is marked as computed
		ID int `tf:"id"`
	}

	// This is the top level struct. All the computed fields should mark here
	type Student struct {
        // id to denotes the resource or datasource. This should be computed as
        // well as string type
		ID    string `tf:"id,computed"`
		Class string `tf:"class,computed"`
		// since subject.#.id is computed the top level field should mark as computed
		Subject []Subject `tf:"name,computed"`
		// sub is used here since we only expects only elements in Mark and it is a set as well
		// You can also use Mark []Mark as well. Both will work
		Mark Mark `tf:"mark,sub,computed"`
	}

	var studentModel Student
    // Parse to student model
	err := tftags.Get(d, &studentModel)
	if err != nil {
		return err
	}

```

ID in the struct will considered as the ID of the resource/data-source and `Id()` will be called if top level `id` tag is found.
Nested structs are supported for nested terraform resources.

_**Note:** Please note that, you need to provide `tf` tags for every fields to parse, any fields without `tf` tags will be omitted_

You can use `Set` function to set computed fields of a resource. To denote computed field use `computed` in the tags.

example:
```go
    if err:= tftags.Set(d,studentModel);err!=nil{
        return err
    }
```
Top level `tf:"id"` with `computed` tag will consider as id of the resource and tftags will call `SetId()` function for the same. This will panics
if the ID type is not `string`

_**Note:** You need to set computed on top level field of the struct, otherwise it won't work_

### where to use `sub` in tf tag
If you have request json like following
```json
{
    "name": "test",
    "data": {
        "id": 1
    }
}
```
and struct model looks like
```go
    type IDModel struct{
        ID int `json:"id" tf:"id,computed"`
    }

    type Model struct{
        ID string `json:"id" tf:"id,computed"`
        Name string  `json:"name" tf:"name"`
        Data IDModel `json:"data" tf:"data,sub,computed"`
    }
```
Since terraform doesn't support child object like json do.So using sub will helps to mock child struct. For using
`sub` the corresponding entry in terraform schema should be `schema.TypeSet` or `schema.TypeList`.

# License
https://github.com/tshihad/tftags/blob/main/LICENSE
