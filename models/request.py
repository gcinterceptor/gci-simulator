

class Request(object):

    def __init__(self, conf, env, id, created_at, load_balancer):
        self.memory = float(conf['memory'])

        self.env = env
        self.id = id
        self.load_balancer = load_balancer
        self.service_time = 0

        self.created_time = created_at  # The time when the request was created
        self._arrived_time = None  # The time when the request arrived at server
        self._latency = None

        self.done = False

    def run(self, heap, service_time):
        self.service_time = service_time
        yield self.env.timeout(service_time)
        yield heap.put(self.memory)
        self.done = True

    def arrived_at(self, arrived_time):
        self._arrived_time = arrived_time

    def finished_at(self, finished_time):
        self._latency = finished_time - self.created_time

