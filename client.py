import Pyro4, time, os, signal, subprocess, sys
import sqlite3
from job import *

Usage = "Usage: "
dbVideoIndex = None

job_server = Pyro4.Proxy("PYRONAME:jobserver.name")

def serverIsUp():
	try:
		ns = Pyro4.locateNS()
	except Pyro4.errors.NamingError:
		return False
	return True

def cleanup():
	nsPid, dPid = job_server.cleanup()
	os.kill(dPid, signal.SIGTERM)
	time.sleep(1)
	os.kill(nsPid, signal.SIGTERM)
	print "Daemon(PID {0}) and NameServer(PID {1}) terminated.".format(dPid, nsPid)

def setup():
	try:
		pid = subprocess.Popen(["python", "jobserver.py"]).pid
	except OSError:
		print "Could not start Server: OSError"
		exit(1)
	except ValueError:
		print "Could not start Server: ValueError"
		exit(1)
	print "Server(PID {0}) started.".format(pid)

if len(sys.argv) > 1:
	if str(sys.argv[1]) == "server":
		if str(sys.argv[2]) == "stop":
			if serverIsUp():
				cleanup()
			else:
				print "Server is down."
			exit()
		elif str(sys.argv[2]) == "start":
			if serverIsUp():
				print "Server is up."
			else:
				setup()
			exit()
		elif str(sys.argv[2]) == "restart":
			if serverIsUp():
				cleanup()
				setup()
			else:
				setup()
			print "Server (re)started"
			exit()
		else:
			print Usage
			exit()
	elif str(sys.argv[1]) == "play":
		if len(sys.argv) == 3:
			dbVideoIndex = int(sys.argv[2])
		else:
			pass
	else:
		print Usage
		exit()
else:
	print Usage
	exit()


if serverIsUp():
	pass
else:
	setup()
	print "Starting NameServer... ",
	while not serverIsUp():
		continue
	print "Done."

conn = sqlite3.connect('playlist.db')
c = conn.cursor()

c.execute("SELECT COUNT(*) FROM playlist")
length = c.fetchall()[0][0]

if dbVideoIndex is None:
	print "{0} videos in playlist".format(length)
	dbVideoIndex = 0

elif dbVideoIndex < 0:
	print "Invalid video index."
	exit(1)

elif dbVideoIndex < length:
	print "Selected video [{0}]".format(dbVideoIndex)
	length = dbVideoIndex+1
else:
	exit(1)
	
for i in range(dbVideoIndex, length):
	t = (i+1, )
	c.execute("SELECT url FROM playlist WHERE ROWID=?", t)
	video_url = c.fetchone()[0]
	print "Processing [{0}] {1}...".format(i, video_url)
	job_server.execute_job(video_url)

conn.close()
exit()
