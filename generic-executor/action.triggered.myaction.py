import os
import sys

# Lets get the first parameter which could potentially be a local file name
methodArg = ""
if len(sys.argv) > 1:
    methodArg = sys.argv[1]

print("This is my genericactionname handler script and I got passed " + methodArg + " as parameter")
print("I also have some env variables, e.g: PID=" + os.getenv('DATA_PROBLEM_PID', "") + ", SHKEPTNCONTEXT=" + os.getenv('SHKEPTNCONTEXT', ""))
print("SOURCE=" + os.getenv('SOURCE',""))
print("PROJECT=" + os.getenv('DATA_PROJECT',""))
print("PROBLEMTITLE=" + os.getenv('DATA_PROBLEM_PROBLEMTITLE',""))
print("And here is the message that was passed as part of the remediation action definition :" + os.getenv("DATA_ACTION_VALUE_MESSAGE", "NO MESSAGE"))