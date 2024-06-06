# Terraform Exec package

## HCL handling

This package is using the `hcl` package to handle the storage of variable's values used.

See also https://hclguide.readthedocs.io/en/latest/go_decoding_gohcl.html

## How to write mapping from tfvars file to Golang struct

- High-level fields should be annotated with `hcl:"<tf_field_name>"`
- Nested structure should be annotated with `cty:"field_name"`

For more information, there is a good explanation on [StackOverflow](https://stackoverflow.com/a/78486469/9332386) why nested structure should be annotated with `cty`

## State and Vars handling

For each apply, if no error occured, the tfvars file will be written so that we can easily retrieve the args which were used.

That will also help to modify easily the configuration of the different resources.
