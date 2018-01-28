package main

import (
  "text/template"
  "encoding/json"
)

var (
  functionTemplates = template.FuncMap{
    "json": func(d interface{}) (string) {
      payload, err := json.Marshal(d)
      if err != nil {
        return ""
      }

      return string(payload[:])
    },
    "json_indent": func(d interface{}) (string) {
      payload, err := json.MarshalIndent(d, "", "  ")
      if err != nil {
        return ""
      }

      return string(payload[:])
    },
  }
)
