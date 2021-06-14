#
#    Copyright (C) 2013,2014 Loic Dachary <loic@dachary.org>
#
#    This program is free software: you can redistribute it and/or modify
#    it under the terms of the GNU Affero General Public License as published by
#    the Free Software Foundation, either version 3 of the License, or
#    (at your option) any later version.
#
#    This program is distributed in the hope that it will be useful,
#    but WITHOUT ANY WARRANTY; without even the implied warranty of
#    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#    GNU Affero General Public License for more details.
#
#    You should have received a copy of the GNU Affero General Public License
#    along with this program.  If not, see <http://www.gnu.org/licenses/>.
#
set -e
set -x
set -u

DIR=${1}

# reset
pkill ceph || true
rm -rf ${DIR}/*
LOG_DIR=${DIR}/log
MON_DATA=${DIR}/mon
MDS_DATA=${DIR}/mds
MOUNTPT=${MDS_DATA}/mnt
OSD_DATA=${DIR}/osd
RGW_DATA=${DIR}/radosgw
mkdir ${LOG_DIR} ${MON_DATA} ${OSD_DATA} ${MDS_DATA} ${MOUNTPT} ${RGW_DATA}
MDS_NAME="Z"
MON_NAME="a"
MGR_NAME="x"
MIRROR_ID="m"
RGW_ID="r"
S3_ACCESS_KEY=2262XNX11FZRR44XWIRD
S3_SECRET_KEY=rmtuS1Uj1bIC08QFYGW18GfSHAbkPqdsuYynNudw

# cluster wide parameters
cat >> ${DIR}/ceph.conf <<EOF
[global]
fsid = $(uuidgen)
osd crush chooseleaf type = 0
run dir = ${DIR}/run
auth cluster required = none
auth service required = none
auth client required = none
osd pool default size = 1
mon host = ${HOSTNAME}

[mds.${MDS_NAME}]
host = ${HOSTNAME}

[mon.${MON_NAME}]
log file = ${LOG_DIR}/mon.log
chdir = ""
mon cluster log file = ${LOG_DIR}/mon-cluster.log
mon data = ${MON_DATA}
mon data avail crit = 0
mon addr = ${HOSTNAME}
mon allow pool delete = true

[osd.0]
log file = ${LOG_DIR}/osd.log
chdir = ""
osd data = ${OSD_DATA}
osd journal = ${OSD_DATA}.journal
osd journal size = 100
osd objectstore = memstore
osd class load list = *
osd class default list = *

[client.rgw.${RGW_ID}]
rgw dns name = ${HOSTNAME}
rgw enable usage log = true
rgw usage log tick interval = 1
rgw usage log flush threshold = 1
rgw usage max shards = 32
rgw usage max user shards = 1
log file = /var/log/ceph/client.rgw.${RGW_ID}.log
rgw frontends = beast port=80
EOF

export CEPH_CONF=${DIR}/ceph.conf

# start an osd
ceph-mon --id ${MON_NAME} --mkfs --keyring /dev/null
touch ${MON_DATA}/keyring
ceph-mon --id ${MON_NAME}

# start an osd
OSD_ID=$(ceph osd create)
ceph osd crush add osd.${OSD_ID} 1 root=default
ceph-osd --id ${OSD_ID} --mkjournal --mkfs
ceph-osd --id ${OSD_ID} || ceph-osd --id ${OSD_ID} || ceph-osd --id ${OSD_ID}

# start an mds for cephfs
ceph auth get-or-create mds.${MDS_NAME} mon 'profile mds' mgr 'profile mds' mds 'allow *' osd 'allow *' > ${MDS_DATA}/keyring
ceph osd pool create cephfs_data 8
ceph osd pool create cephfs_metadata 8
ceph fs new cephfs cephfs_metadata cephfs_data
ceph fs ls
ceph-mds -i ${MDS_NAME}
ceph status
while [[ ! $(ceph mds stat | grep "up:active") ]]; do sleep 1; done


# start a manager
ceph-mgr --id ${MGR_NAME}

# start rbd-mirror
ceph auth get-or-create client.rbd-mirror.${MIRROR_ID} mon 'profile rbd-mirror' osd 'profile rbd'
rbd-mirror --id ${MIRROR_ID} --log-file ${LOG_DIR}/rbd-mirror.log

# start cephfs-mirror
# Skip on "nautilus" and "octopus" the supported ceph versions that
# don't have it.
case "${CEPH_VERSION}" in
    nautilus|octopus)
        echo "Skipping cephfs-mirror on ${CEPH_VERSION} ..."
    ;;
    *)
        ceph auth get-or-create "client.cephfs-mirror.${MIRROR_ID}" \
            mon 'profile cephfs-mirror' \
            mds 'allow r' \
            osd 'allow rw tag cephfs metadata=*, allow r tag cephfs data=*' \
            mgr 'allow r'
        cephfs-mirror --id "cephfs-mirror.${MIRROR_ID}" \
            --log-file "${LOG_DIR}/cephfs-mirror.log"
        ceph fs authorize cephfs client.cephfs-mirror-remote / rwps > "${DIR}/cephfs-mirror-remote.out"
        # the .out file above is not used by the scripts but can be used for debugging
    ;;
esac


# start an rgw
ceph auth get-or-create client.rgw."${RGW_ID}" osd 'allow rwx' mon 'allow rw' -o ${RGW_DATA}/keyring
radosgw -n client.rgw."${RGW_ID}" -k ${RGW_DATA}/keyring
timeout 60 sh -c 'until [ $(ceph -s | grep -c "rgw:") -eq 1 ]; do echo "waiting for rgw to show up" && sleep 1; done'
radosgw-admin user create --uid admin --display-name "Admin User" --caps "buckets=*;users=*;usage=read;metadata=read" --access-key="$S3_ACCESS_KEY" --secret-key="$S3_SECRET_KEY"

# test the setup
ceph --version
ceph status
test_pool=$(uuidgen)
temp_file=$(mktemp)
ceph osd pool create ${test_pool} 0
rados --pool ${test_pool} put group /etc/group
rados --pool ${test_pool} get group ${temp_file}
diff /etc/group ${temp_file}
ceph osd pool delete ${test_pool} ${test_pool} --yes-i-really-really-mean-it
rm ${temp_file}

touch ${DIR}/.ready

# vim: set ts=4 sw=4 sts=4 et:
