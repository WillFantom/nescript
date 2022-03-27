#!/bin/bash
set -e

echo "{{.Title}} Example"
{{range .People}}    
echo -e "\tName: {{.Name.First}} {{.Name.Last}}"
echo -e "\tCharacter: {{.Character}}"
echo -e "\t-----"
{{end}}

echo "IMDB Score: "$IMDB
