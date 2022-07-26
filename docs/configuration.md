Configuration
=============

Editing Config
--------------

Aretext stores its configuration in a single YAML file. You can edit the config file using the `-editconfig` flag:

```
aretext -editconfig
```

The configuration file is located at `$XDG_CONFIG_HOME/aretext/config.yaml`, where `XDG_CONFIG_HOME` is configured according to the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html). On Linux, this defaults to `~/.config`, and on macOS it defaults to `~/Library/Application Support`.

When you open the config file, you should see something like:

```yaml
- name: default
  pattern: "**"
  config:
    autoIndent: false
    hideDirectories: [".git"]
    syntaxLanguage: plaintext
    tabExpand: false
    tabSize: 4
    showLineNumbers: false

- name: json
  pattern: "**/*.json"
  config:
    autoIndent: true
    syntaxLanguage: json
    tabExpand: true
    tabSize: 2
    showLineNumbers: true
```

Each item in the configuration file describes a *rule*. For example, in the snippet above, the first rule is named "default" and the second rule is named "json".

Each rule has a *pattern*. The "\*\*" is a wildcard that matches any subdirectory, and "\*" is a wildcard that matches zero or more characters in a file or directory name.

When aretext loads a file, it checks each rule in order. If the rule's pattern matches the file's absolute path, it applies the rule to update the configuration.

For example, if aretext loaded the file "foo/bar.json" using the above configuration, both rules would match the filename. The resulting configuration would be:

```yaml
config:
  autoIndent: true           # from the "json" rule
  hideDirectories: [".git"]  # from the "default" rule
  syntaxLanguage: json       # from the "json" rule
  tabExpand: true            # from the "json" rule
  tabSize: 2                 # from the "json" rule
  showLineNumbers: true      # from the "json" rule
```

When merging configurations from different rules:

-	For strings and numbers, the values from later rules overwrite the values from previous rules.
-	For lists, the values from all rules are combined.
-	For dictionaries, the keys from later rules are added to the merged dictionary, potentially overwriting keys set by previous rules.

This is a powerful mechanism for customizing configuration based on filename extension and/or project location. For example, suppose that one project you work on uses four spaces to indent JSON files. You could add a new rule to your config that overwrites the tabSize for JSON files in that specific project:

```yaml
# ... other rules above ...
- name: myproject json
  pattern: "**/myproject/**/*.json"
  config:
    tabSize: 4
```

Troubleshooting
---------------

### Fixing errors on startup

If your YAML config file has errors, aretext will exit with an error message. You can force aretext to ignore the config file by passing the "-noconfig" flag:

```
aretext -editconfig -noconfig
```

This allows you to start the editor so you can fix the configuration.

### Checking which rules were applied

To see which configuration rules aretext applied when loading a file, start aretext with logging enabled:

```
aretext -log debug.log
```

If you view the file `debug.log`, you should see lines like this:

```
Applying config rule 'default' with pattern '**' for path 'path/to/file.txt'
```

This tells you which rules aretext applied when opening a file, which can help you debug your configuration.

Configuration Reference
-----------------------

For a complete list of available configuration options, see [Configuration Reference](config-reference.md).
