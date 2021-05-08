package project

import "text/template"

func LoadTemplates(dir string, unknown bool) (*template.Template, error) {
	// TODO: If unknown is true then use embedded templates "templates/%project%"
	//    if they are providing custom templates then load from os
	return nil, nil
}
