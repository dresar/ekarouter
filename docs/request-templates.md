# Request Templates Engine

Request templates define reusable parameterized HTTP blueprints with safe variable substitution.

## Variable Interpolation
Templates use double-brace syntax: `{{variable_name}}`.

Example URL Template:
```text
https://api.github.com/repos/{{owner}}/{{repo}}/issues
```

## Security Guarantees
- Unset variables remain untouched in the string without causing panic or evaluation errors.
- Header injection prevention: Header names and values are sanitized to strip any carriage return (`\r`) or newline (`\n`) characters, preventing CRLF injection and HTTP response splitting.
