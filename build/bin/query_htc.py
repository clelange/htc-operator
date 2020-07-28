#!/usr/bin/env python
from __future__ import print_function
import sys
import argparse
import json
import subprocess
import htcondor


def transfer(cluster_ids):
    pass


def remove(cluster_ids):
    pass


def get_dict(query, attr_list):
    attr_default = {
        "ExitCode": -1,
        "RemoveReason": ""
    }
    cluster_dict = {}
    for attr in attr_list:
        if attr not in query:
            cluster_dict[attr] = attr_default[attr]
        else:
            cluster_dict[attr] = query[attr]
    return cluster_dict


def query(cluster_ids):
    schedd = htcondor.Schedd()
    attr_list = ["ClusterId", "ProcId", "JobStatus", "EnteredCurrentStatus", "ExitCode", "RemoveReason"]
    status = []

    for cluster_id in cluster_ids:
        query = schedd.query(
                    constraint='ClusterId=?={}'.format(cluster_id),
                    attr_list=attr_list)
        if query:
            for query_item in query:
                cluster_dict = get_dict(query_item, attr_list)
                status.append(cluster_dict)
        else:
            condor_it = schedd.history('ClusterId == {}'.format(cluster_id), attr_list, match=1)
            for query_item in condor_it:
                if query_item:
                    cluster_dict = get_dict(query_item, attr_list)
                    status.append(cluster_dict)
    print(json.dumps(status))


def main():

    parser = argparse.ArgumentParser(
                description='Manipulate, submit and query HTCondor jobs.')
    parser.add_argument('cluster_ids', metavar='891629', type=int, nargs='+',
                        help='List of ClusterIds')
    query_type = parser.add_mutually_exclusive_group(required=True)
    query_type.add_argument('--transfer', '-t', action='store_true',
                            default=False,
                            help='Transfer output of ClusterIds')
    query_type.add_argument('--remove', '-r', action='store_true',
                            default=False,
                            help='Remove jobs')
    query_type.add_argument('--status', '-s', action='store_true',
                            default=False,
                            help='Query status')
    args = parser.parse_args()

    if args.transfer:
        transfer(args.cluster_ids)
    if args.status:
        query(args.cluster_ids)
    if args.status:
        remove(args.cluster_ids)






if __name__ == "__main__":
    main()
