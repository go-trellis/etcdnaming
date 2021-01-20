/*
Copyright © 2017 Henry Huang <hhh@rutcode.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package etcdnaming

import "time"

// ServerRegister regist or revoke server interface
type ServerRegister interface {
	// Regist 注册服务
	Regist() error
	// Revoke 注销服务
	Revoke() error
}

// ServerRegisterConfig struct server regiter config
// name: server name
// target: etcd' client url, separate by ','
// serv: server address host:port
// interval: Rotation time to registe serv into etcd
// ttl: expired time, seconds
// registRetryTimes: allow failed to regist server and retry times; -1 alaways retry
type ServerRegisterConfig struct {
	Name             string
	Target           string
	Service          string // host & port
	Version          string // server's version
	TTL              time.Duration
	Interval         time.Duration
	RegistRetryTimes int

	serverName string // name & version
}
