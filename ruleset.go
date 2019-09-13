/*
 * Copyright 2018 - 2019 Christian Müller <dev@c-mueller.xyz>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ads

import "regexp"

type UpdateableRuleset struct {
	Blacklist      map[string]bool
	Whitelist      map[string]bool
	BlacklistSources []string
	WhitelistSources []string
}

type ConfiguredRuleSet struct {
	Blacklist      map[string]bool
	Whitelist      map[string]bool
	WhitelistRegex []*regexp.Regexp
	BlacklistRegex []*regexp.Regexp
}

func BuildRuleset(whitelist, blacklist []string) ConfiguredRuleSet {
	r := ConfiguredRuleSet{
		Blacklist:      make(map[string]bool),
		Whitelist:      make(map[string]bool),
		WhitelistRegex: make([]*regexp.Regexp, 0),
		BlacklistRegex: make([]*regexp.Regexp, 0),
	}

	for _, v := range whitelist {
		r.AddToWhitelist(v)
	}

	for _, v := range blacklist {
		r.AddToBlacklist(v)
	}

	return r
}


func (r *ConfiguredRuleSet) AddRegexToWhitelist(regex string) error {
	exp, err := regexp.Compile(regex)
	if err != nil {
		return err
	}

	r.WhitelistRegex = append(r.WhitelistRegex, exp)

	return nil
}

func (r *ConfiguredRuleSet) AddRegexToBlacklist(regex string) error {
	exp, err := regexp.Compile(regex)
	if err != nil {
		return err
	}

	r.BlacklistRegex = append(r.BlacklistRegex, exp)

	return nil
}

func (r *ConfiguredRuleSet) AddToWhitelist(qname string) {
	r.Whitelist[qname] = true
}

func (r *ConfiguredRuleSet) AddToBlacklist(qname string) {
	r.Blacklist[qname] = true
}

func (r *ConfiguredRuleSet) IsWhitelisted(qname string) bool {
	for _, v := range r.WhitelistRegex {
		if v.MatchString(qname) {
			return true
		}
	}

	return r.Whitelist[qname]
}

func (r *ConfiguredRuleSet) IsBlacklisted(qname string) bool {
	for _, v := range r.BlacklistRegex {
		if v.MatchString(qname) {
			return true
		}
	}
	return r.Blacklist[qname]
}
