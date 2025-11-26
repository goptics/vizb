```bash
vizb bash/decode-encode.txt -r '(?<n>[^/]+)/text=(?<x>[^/]+)/level=(?<y>.*)' -o output/out-decode

vizb bash/validator.txt -r '(?<y>.+?)(?:(?<x>Success|Failure)(?<n>.*)|$)' -o output/validator

vizb bash/hash.txt -r 'Hashing64(?<x>.*)' -o output/hash

vizb bash/marshal.txt -r '(?<x>.*)By(?<y>.*\)' -o output/marshal
```
