FROM centos:centos7
ADD ./data_collection_tree /data_collection_tree
ENTRYPOINT ["/data_collection_tree"]
