from log import get_logger
import simpy

class LoadBalancer(object):

    def __init__(self, env, conf, communication_time, log_path=None):
        self.env = env
        
        self.communication_time = communication_time
        
        self.actual_server = 0
        self.servers = list()
        
        if log_path:
            self.logger = get_logger(log_path + "/loadbalancer.log", "LOAD BALANCER")
        else:
            self.logger = None

    def get_next_server(self):
        self.actual_server = (self.actual_server + 1) % len(self.servers) 
        return self.actual_server
            
    def add_server(self, server):
        self.servers.append(server)

    def request_arrived(self, request):
        if self.logger:
            self.logger.info(" At %.3f, request %d arrived" % (self.env.now, request.id))
        
        request.sent_at()
        
        server = self.get_next_server()
        self.env.process(self.send_to(server, request))
    
    def send_to(self, server, request):
        request.redirected()
        
        if self.logger:
            self.logger.info(" At %.3f, request %d was send to server %d" % (self.env.now, request.id, self.servers[server].id))
        
        yield self.env.timeout(self.communication_time)
        
        self.servers[server].request_arrived(request)

    def success_request(self, request):
        yield self.env.timeout(self.communication_time)
        
        if self.logger:
            self.logger.info(" At %.3f, request %d processed" % (self.env.now, request.id))
            
        request.client.success_request(request)

    def shed_request(self, request, server):
        yield self.env.timeout(self.communication_time)
        
        if self.logger:
            self.logger.info(" At %.3f, Request %d was shedded" % (self.env.now, request.id))
        
        if request.redirects == len(self.servers):
            if self.logger:
                self.logger.info(" At %.3f, request %d was refused" % (self.env.now, request.id))
            
            request.client.refuse_request(request)
            
        else:
            server = self.get_next_server()
            self.env.process(self.send_to(server, request))
