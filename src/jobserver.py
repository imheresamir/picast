import Pyro4, daemon, subprocess, os, sys, time
from job import Job

class JobServer:
  def __init__(self, nsPid, dPid):
    self.nsPid = nsPid
    self.dPid = dPid
    self.currentDir = os.environ['PWD']
    
    self.isDownloading = 0
    self.isPlaying = 0
    self.currentJob = None
  
  # @Pyro4.oneway ## Controls if Pyro4 comms are blocking or not
  def execute_job(self, video_url):
    self.currentJob = Job(video_url)
    
    dlCode = self._download()
    if dlCode != 0:
      print "Server: DownloadProcess ErrorCode {0}".format(dlCode)
    else:
      self.isDownloading = 0

    
    playCode = self._play()
    if playCode is None:
      print "Server: PlayProcess UserInterrupt"
    elif playCode != 0:
      print "Server: PlayProcess ErrorCode {0}".format(playCode)
    else:
      self.isPlaying = 0

    
    #self.cleanup() # Remove tmp files every time for now
    print "Server: JobExecute SuccessCode 0"
  
  def _download(self):
    self.isDownloading = 1
    self.dlPid = subprocess.Popen(["youtube-dl", "--no-part", "-v", "-o", self.currentJob.outFilename, self.currentJob.video_url], cwd=self.currentDir).pid
    #return dlProcess.wait()
    while not os.path.exists(self.currentDir + "/" + self.currentJob.outFilename):
      time.sleep(1)
    return 1
  
  def _play(self):
    self.isPlaying = 1
    self.playerPid = subprocess.Popen(["omxplayer", self.currentJob.outFilename], cwd=self.currentDir).pid
    #playerProcess.wait()
    self.isPlaying = 0

  def cleanup(self):
    if(self.currentJob):
      rmProcess = subprocess.Popen(["rm", self.currentJob.outFilename], cwd=self.currentDir)
      rmProcess.wait()
      self.currentJob = None
    else:
      pass
    return (self.nsPid, self.dPid)

d1 = daemon.DaemonContext()
#d1.stdin = sys.stdin
log_file = open('Server.log', 'w')
d1.stdout = log_file
#d1.stdout = sys.stdout

with d1:
  nsPid = subprocess.Popen("pyro4-ns").pid
  
  dPid = os.getpid()
  
  job_server = JobServer(nsPid, dPid)
  d = Pyro4.Daemon()

  while True:
    try:
      ns = Pyro4.locateNS()
      #print "got the ns"
      break
    except Pyro4.errors.NamingError:
      #print "can't get ns!"
      #print "nsPid: {0}".format(nsPid)
      exit(1)

  uri = d.register(job_server)
  ns.register("jobserver.name", uri)
  
  #print "jobserver registered"
  d.requestLoop()
