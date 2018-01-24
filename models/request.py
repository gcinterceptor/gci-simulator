from log import get_logger


class Request(object):

    def __init__(self, id, env, created_at, load_balancer, conf, log_path=None):
        self.id = id
        self.env = env
        self.load_balancer = load_balancer
        self.service_time = float(conf['service_time'])
        self.memory = float(conf['memory'])

        self.done = False

        self.created_time = created_at  # The time when the request was created
        self._arrived_time = None  # The time when the request arrived at server
        self._attended_time = None  # The time when the request was taken out the queue at server
        self._finished_time = None  # The time when the request was finished at server

        self._latency = None

        self._time_in_queue = None
        self._time_in_server = None

        if log_path:
            self.logger = get_logger(log_path + "/request.log", "REQUEST")
        else:
            self.logger = None

    def run(self, heap):
        self._time_in_queue = self.env.now - self._arrived_time
        yield self.env.timeout(self.service_time)
        yield heap.put(self.memory)
        self.done = True

    def arrived_at(self, time):
        self._arrived_time = time

    def attended_at(self, time):
        self._attended_time = time
        self._time_in_server = self._attended_time - self._arrived_time

    def finished_at(self, time):
        self._finished_time = time
        self._latency = self._finished_time - self.created_time
        if self.logger:
            self.logger.info(" At %.3f, Request %i was finished. Latency: %.3f" % (self.id, self._finished_time, self._latency))

