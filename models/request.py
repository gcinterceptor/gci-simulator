

class Request(object):

    def __init__(self, env, id, created_at, load_balancer):
        self.env = env
        self.id = id
        self.load_balancer = load_balancer
        self.service_time = 0

        self.created_time = created_at  # The moment when the request was created
        self.times_forwarded = 0
        self._arrived_time = None  # The moment when the request arrived at server
        self._latency = None

        self.done = False

    def run(self, service_time):
        self.service_time = service_time
        yield self.env.timeout(service_time)
        self.done = True

    def arrived_at(self, arrived_time):
        self._arrived_time = arrived_time

    def finished_at(self, finished_time):
        self._latency = finished_time - self.created_time

