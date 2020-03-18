import sys
import htcondor
lines = list(map(lambda x: x.split(' '), open(sys.argv[1], 'r').read().split("\n")[:-1]))
procList = list(map(lambda x: "(ClusterId == %s && ProcId == %s)" % tuple(x[0].split('.')), lines))

schedd = htcondor.Schedd()
queryResult = schedd.query(attr_list=["ClusterId", "ProcId", "JobStatus"],
    constraint=reduce(lambda a, b: a + " || " + b, procList))
for i in range(0, len(lines)):
    (origClus, origProc) = lines[i][0].split('.')
    for currQ in queryResult:
        if (str(currQ["ClusterId"]) == origClus) and (str(currQ["ProcId"]) == origProc):
            lines[i].append(str(currQ["JobStatus"]))

for l in lines:
    print(reduce(lambda a, b: a + " " + b, l))
