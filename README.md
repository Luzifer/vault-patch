# Luzifer / vault-patch

`vault-patch` is a very small utility to "patch" a [Vault](https://vaultproject.io/) key. In case you don't want to replace the whole data stored in that key but only want to modify one (or more) specific data pair(s) this can come handy:

```bash
# vault read secret/test
Key                     Value
---                     -----
refresh_interval        768h0m0s
field1                  test1
field2                  test2


# vault-patch secret/test field2=test4 field3=test3
INFO[0000] Data successfully written to key "secret/test"

# vault read secret/test
Key                     Value
---                     -----
refresh_interval        768h0m0s
field1                  test1
field2                  test4
field3                  test3
```

As you can see only the data given in the command was touched and the `field1` was kept as it was before.
