# redirect

A simple service for redirecting HTTP requests

## Usage

**Container Image**

`guoyk/redirect`

**Environment Variables**

* `REDIRECT_LISTEN`, listen address, default to `:80`
* `REDIRECT_TARGET`, redirect target, if ends with `/`, source request path will be appended to `REDIRECT_TARGET`
* `REDIRECT_PREFIX`, works when `REDIRECT_TARGET` has tailing `/`, trim prefix from original url
* `REDIRECT_PERMANENT`, use `301` instead of `302` for redirecting
* `REDIRECT_VERBOSE`, set to `true` to enable verbose logging

## Example

* When `REDIRECT_TARGET=http://b.example.com`

    ```
    http://a.example.com/ -> http://b.example.com
    http://a.example.com/aaa -> http://b.example.com
    ```

* When `REDIRECT_TARGET=http://b.example.com/`

  ```
  http://a.example.com/ -> http://b.example.com/
  http://a.example.com/aaa -> http://b.example.com/aaa
  ```

* When `REDIRECT_TARGET=http://b.example.com/ccc/` and `REDIRECT_PREFIX=/aaa`

  ```
  http://a.example.com/aaa/bbb -> http://b.example.com/ccc/bbb
  ```

## Donation

Check https://guoyk.xyz/donation

## Credits

Guo Y.K., MIT License
