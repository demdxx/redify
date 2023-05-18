package appcontext

import (
	"os"
	"regexp"
	"strings"
)

var regExpEnvVarsExpression = regexp.MustCompile(`\$\{\{(\s*[a-zA-Z0-9_]+\s*)\}\}`)

func (c *ConfigType) Prepare() {
	c.Cache.Connect = prepareItem(c.Cache.Connect)
	for i := range c.Sources {
		src := &c.Sources[i]
		src.Connect = prepareItem(c.Sources[i].Connect)
		src.NotifyChannel = prepareItem(c.Sources[i].NotifyChannel)
		for j := range src.Binds {
			bind := &src.Binds[j]
			bind.TableName = prepareItem(bind.TableName)
			bind.Key = prepareItem(bind.Key)
			bind.WhereExt = prepareItem(bind.WhereExt)
			bind.GetQuery = prepareItem(bind.GetQuery)
			bind.ListQuery = prepareItem(bind.ListQuery)
			bind.UpsertQuery = prepareItem(bind.UpsertQuery)
			bind.DelQuery = prepareItem(bind.DelQuery)
			for k := range bind.DatatypeMapping {
				dm := &bind.DatatypeMapping[k]
				dm.Name = prepareItem(dm.Name)
				dm.Type = prepareItem(dm.Type)
			}
		}
	}
}

func prepareItem(s string) string {
	return regExpEnvVarsExpression.ReplaceAllStringFunc(s, func(s string) string {
		envName := strings.TrimPrefix(s, "${{")
		envName = strings.TrimSuffix(envName, "}}")
		envName = strings.TrimSpace(envName)
		v, _ := os.LookupEnv(envName)
		return v
	})
}
