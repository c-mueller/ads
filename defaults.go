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

import (
	"net"
	"time"
)

var defaultBlacklists = []string{
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
	"https://mirror1.malwaredomains.com/files/justdomains",
	"http://sysctl.org/cameleon/hosts",
	"https://zeustracker.abuse.ch/blocklist.php?download=domainblocklist",
	"https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt",
	"https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
	"https://hosts-file.net/ad_servers.txt",
}

var strictDefaultBlacklists = []string{
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts",
	"https://mirror1.malwaredomains.com/files/justdomains",
	"http://sysctl.org/cameleon/hosts",
	"https://zeustracker.abuse.ch/blocklist.php?download=domainblocklist",
	"https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt",
	"https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
	"https://hosts-file.net/ad_servers.txt",
	"https://hosts-file.net/grm.txt",
	"https://reddestdream.github.io/Projects/MinimalHosts/etc/MinimalHostsBlocker/minimalhosts",
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/data/KADhosts/hosts",
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/data/add.Spam/hosts",
	"https://v.firebog.net/hosts/static/w3kbl.txt",
	"https://v.firebog.net/hosts/BillStearns.txt",
	"https://www.dshield.org/feeds/suspiciousdomains_Low.txt",
	"https://www.dshield.org/feeds/suspiciousdomains_Medium.txt",
	"https://www.dshield.org/feeds/suspiciousdomains_High.txt",
	"https://www.joewein.net/dl/bl/dom-bl-base.txt",
	"https://raw.githubusercontent.com/matomo-org/referrer-spam-blacklist/master/spammers.txt",
	"https://hostsfile.org/Downloads/hosts.txt",
	"https://someonewhocares.org/hosts/zero/hosts",
	"https://raw.githubusercontent.com/Dawsey21/Lists/master/main-blacklist.txt",
	"https://raw.githubusercontent.com/vokins/yhosts/master/hosts",
	"https://hostsfile.mine.nu/hosts0.txt",
	"https://v.firebog.net/hosts/Kowabit.txt",
	"https://adaway.org/hosts.txt",
	"https://v.firebog.net/hosts/AdguardDNS.txt",
	"https://raw.githubusercontent.com/anudeepND/blacklist/master/adservers.txt",
	"https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt",
	"https://hosts-file.net/ad_servers.txt",
	"https://v.firebog.net/hosts/Easylist.txt",
	"https://pgl.yoyo.org/adservers/serverlist.php?hostformat=hosts;showintro=0",
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/data/UncheckyAds/hosts",
	"https://www.squidblacklist.org/downloads/dg-ads.acl",
	"https://v.firebog.net/hosts/Easyprivacy.txt",
	"https://v.firebog.net/hosts/Prigent-Ads.txt",
	"https://gitlab.com/quidsup/notrack-blocklists/raw/master/notrack-blocklist.txt",
	"https://raw.githubusercontent.com/StevenBlack/hosts/master/data/add.2o7Net/hosts",
	"https://raw.githubusercontent.com/crazy-max/WindowsSpyBlocker/master/data/hosts/spy.txt",
	"https://zerodot1.gitlab.io/CoinBlockerLists/hosts",
}

const defaultIPv4ResolutionIP = "127.0.0.1"
const defaultIPv6ResolutionIP = "::1"

var defaultConfigWithoutRules = adsPluginConfig{
	BlacklistURLs: []string{},
	WhitelistURLs: []string{},

	BlacklistFiles: []string{},
	WhitelistFiles: []string{},

	BlacklistRules: []string{},
	WhitelistRules: []string{},

	RegexBlacklistRules: []string{},
	RegexWhitelistRules: []string{},

	TargetIP:   net.ParseIP(defaultIPv4ResolutionIP),
	TargetIPv6: net.ParseIP(defaultIPv6ResolutionIP),

	HttpListRenewalInterval:  time.Hour * 24,
	FileListRenewalInterval:  time.Minute,
	ListRenewalRetryCount:    5,
	ListRenewalRetryInterval: time.Minute,

	ListPersistencePath:   "",
	EnableLogging:         false,
	EnableAutoUpdate:      true,
	EnableListPersistence: false,
	WriteNXDomain:         false,
}
