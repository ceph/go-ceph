FROM go-ceph-tmp

COPY micro-osd.sh /
COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
