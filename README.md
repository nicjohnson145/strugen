# strugen

```
pronounced strew-gen
```

Base library for writing code generators that operate on structs and their fields
`strugen` is designed to be consumed by code used in `//go:generate`. 

## Basic Usage

Given the following generation code, compiled as `access`

```go
package main

import (
	"github.com/spf13/cobra"
	"github.com/nicjohnson145/strugen"
	"fmt"
	"encoding/json"
)

func main() {
    cmd().Execute()
}

func cmd() *cobra.Command {
	var types []string
	root := &cobra.Command{
		Use: "access",
		Run: func(cmd *cobra.Command, args []string) {
			run(types)
		},
	}
	root.Flags().StringSliceVarP(&types, "type", "t", []string{}, "Types to generate for")
	return root
}


func run(types []string) {
	gen := strugen.Generator{
		Types: types,
		TagName: "access",
	}
	structs, err := gen.FindStructs()
	if err != nil {
        panic(err)
	}

	bytes, err := json.MarshalIndent(structs, "", "   ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))
}
```

and the following (separate) repo using said binary as 

```go
package main

//go:generate access -t BlargType -t SmartType

type FooType struct {
	X string
}

type BlargType struct {
	V1 string  `access:"write"`
	V2 int     `access:"read,write"`
	V3 FooType `access:""`
}

type SmartType struct {
	X1 string
	X2 bool
}
```

will result in the following output

```
{
   "BlargType": {
      "Name": "BlargType",
      "Fields": {
         "V1": {
            "Name": "V1",
            "Exported": "",
            "Tagged": true,
            "TagValue": "write",
            "Type": "string"
         },
         "V2": {
            "Name": "V2",
            "Exported": "",
            "Tagged": true,
            "TagValue": "read,write",
            "Type": "int"
         },
         "V3": {
            "Name": "V3",
            "Exported": "",
            "Tagged": true,
            "TagValue": "",
            "Type": "FooType"
         }
      }
   },
   "SmartType": {
      "Name": "SmartType",
      "Fields": {
         "X1": {
            "Name": "X1",
            "Exported": "",
            "Tagged": false,
            "TagValue": "",
            "Type": "string"
         },
         "X2": {
            "Name": "X2",
            "Exported": "",
            "Tagged": false,
            "TagValue": "",
            "Type": "bool"
         }
      }
   }
}
```
