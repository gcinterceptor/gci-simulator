from log import get_logger
import simpy

class LoadBalancer(object):

    def __init__(self, env, conf, log_path=None):
        self.env = env
        self.sleep = float(conf['sleep_time'])

        self.servers = list()
        self.server_availability = {}
        
        self.actual_server = 0

        if log_path:
            self.logger = get_logger(log_path + "/loadbalancer.log", "LOAD BALANCER")
        else:
            self.logger = None

    def get_next_server(self):
        self.actual_server = (self.actual_server + 1) % len(self.servers) 
        return self.actual_server
            
    def add_server(self, server):
        self.servers.append(server)
        self.server_availability[server.id] = 0

    def request_arrived(self, request):
        if self.logger:
            self.logger.info(" At %.3f, request %d arrived" % (self.env.now, request.id))
        
        yield self.env.process(request.sent_at())
        
        server = self.get_next_server()
        yield self.env.process(self.send_to(server, request))
    
    def send_to(self, server, request):
        yield self.env.process(request.redirected())
        
        if self.logger:
            self.logger.info(" At %.3f, request %d was send to server %d" % (self.env.now, request.id, self.servers[server].id))
        
        yield self.env.process(self.servers[server].request_arrived(request))

    def success_request(self, request):
        if self.logger:
            self.logger.info(" At %.3f, request %d processed" % (self.env.now, request.id))
            
        yield self.env.process(request.client.success_request(request))

    def shed_request(self, request, server, unavailable_time):
        if self.logger:
            self.logger.info(" At %.3f, Request was shedded. The server will be unavailable for: %.3f" % (self.env.now, unavailable_time))
        
        self.server_availability[server.id] = self.env.now + unavailable_time
        
        server = self.get_next_server()
        yield self.env.process(self.send_to(server, request))
