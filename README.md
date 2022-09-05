# ximager

``` sh
# generate private key
openssl genrsa -out private.pem 2048
# extract public key from it
openssl rsa -in private.pem -pubout > public.pem
```
