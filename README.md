# Taskmaster



## How does it work

### Configuration file

The configuration file should be in the `ini` format with the following elements:
    -  

The file is validated and parsed with the [following regex](https://regex101.com/r/ZK8tNM/1):

```regex
/^\s*(?:;|#).*$|^\s*\[\s*([\w _-]+?)\s*\]\s*?(?:(?:;|#).*)?$|^\s*([^=#;\s]+)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^;\s=#]+))\s*?(?:(?:;|#).*)?$/gm
```

- `^\s*(?:;|#).*$`: This first block matches the comment lines
  - Start line with any number of whitespace characters
  - `;` or `#` once
  - Any number of any character until end of line
- `^\s*\[\s*([\w _-]+?)\s*\]\s*?(?:(?:;|#).*)?$`: This section matches the section headers (including inline comments)
  - Start line with any number of whitespace characters
  - `[` once
  - Any number of whitespace characters
  - At least one character of a charset (alnum + underscore + dash + space)
    - This will match as few as character as possible, meaning it won't match surrounding whitespace characters
  - Any number of whitespace characters
  - `]` once
  - Any number of whitespace characters but as few as possible
  - `(?:(?:;|#).*)?`: One or zero inline comment (almost identical as the line comment)
  - End of line
- `^\s*([^=#;\s]+)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^;\s=#]+))\s*?(?:(?:;|#).*)?$`: This section matches the keyval pairs (including inline comments)
  - Start line with any number of whitespace characters
  - At least one character that is not of a charset (comment symbols and `=`)
  - Any number of whitespace characters
  - `=` character
  - Any number of whitespace characters
  - Either one of the three:
    - `"([^"]*)"`: Anything that is not a double quote and between double quotes
    - `'([^']*)'`: Anything that is not a simple quote and between double quotes
    - `([^;\s=#]+)`: At least one character that is not of a charset (comment symbols and `=`)
  - Any number of whitespace characters but as few as possible
  - `(?:(?:;|#).*)?`: One or zero inline comment (almost identical as the line comment)
