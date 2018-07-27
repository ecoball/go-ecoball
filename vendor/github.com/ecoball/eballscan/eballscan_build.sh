#!/bin/bash
##########################################################################
# Copyright 2018 The eballscan Authors
# This file is part of the eballscan.
#
# The eballscan is free software: you can redistribute it and/or modify
# it under the terms of the GNU Lesser General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# The eballscan is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU Lesser General Public License for more details.
#
# You should have received a copy of the GNU Lesser General Public License
# along with the eballscan. If not, see <http://www.gnu.org/licenses/>.
############################################################################

#install cockroachdb
#wget -qO- https://binaries.cockroachdb.com/cockroach-v2.0.4.linux-amd64.tgz | tar  xvz
#sudo cp -i cockroach-v2.0.4.linux-amd64/cockroach /usr/local/bin

#start cockroachdb
cockroach start --insecure --http-port=8081 --background

#create user eballscan
cockroach user set eballscan --insecure

#create databas blockchain
cockroach sql --insecure -e 'create database blockchain'

#grant eballscan
cockroach sql --insecure -e 'GRANT ALL ON DATABASE blockchain TO eballscan'

#build project
make