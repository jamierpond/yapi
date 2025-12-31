# jq variable binding with `as` not supported in jq_filter

## Description

When using the `jq_filter` field in yapi files, standard jq variable binding syntax using `as` doesn't work. This is a limitation compared to standalone jq.

## Steps to Reproduce

Create a yapi file with a jq_filter that uses variable binding:

```yaml
yapi: v1
url: https://example.com/api/graphql
method: POST

graphql: |
  {
    products {
      edges {
        node {
          title
          variants {
            edges {
              node {
                title
                price
              }
            }
          }
        }
      }
    }
  }

jq_filter: |
  .data.products.edges[] |
  (.node.title) as $productTitle |
  .node.variants.edges[] | {
    product: $productTitle,
    variant: .node.title,
    price: .node.price
  }
```

## Expected Behavior

The jq_filter should work as it does in standalone jq, binding `.node.title` to the variable `$productTitle` for use in nested iterations.

## Actual Behavior

yapi returns validation errors:
```
[ERROR] JQ syntax error: unexpected token "|"
[ERROR] variable 'productTitle' is not defined in any environment or defaults
```

## Workaround

Restructure the jq_filter to avoid variable binding by using nested array construction:

```yaml
jq_filter: |
  [.data.products.edges[] |
    {product: .node.title, variants: [.node.variants.edges[] |
      {title: .node.title, price: .node.price}
    ]}
  ]
```

## Impact

This limitation makes it harder to write clean jq filters when you need to reference parent object fields in nested iterations. The workaround requires restructuring data or post-processing.

## Environment

- yapi version: (current from weloveraw-landing project)
- OS: macOS (Darwin 24.6.0)

## Notes

This might be a limitation of the jq implementation used in yapi rather than a bug. If this is intentional, it would be helpful to document which jq features are supported in the jq_filter field.
