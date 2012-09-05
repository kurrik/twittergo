# Copyright 2012 Arne Roomann-Kurrik
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

PROJROOT=$GOPATH/src/github.com/kurrik
PROJNAME=twittergo
PROJPATH=$PROJROOT/$PROJNAME

function yellow {
  echo -e "\033[1;33m$1\033[0m $2"
}

function red {
  echo -e "\033[1;31m$1\033[0m $2"
}

if [ -z "$GOPATH" ]; then
  red "Empty \$GOPATH!"
  exit 1
fi
  
if [ -e $PROJPATH ]; then
  yellow "Removing $PROJPATH"
  rm -rf $PROJPATH
fi

yellow "Ensuring $PROJROOT is created"
mkdir -p $PROJROOT

yellow "Creating link"
ln -s `pwd` $PROJPATH
