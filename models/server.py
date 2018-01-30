from random import uniform
import simpy


class ServerBaseline(object):

    def __init__(self, conf, env, id):
        self.baseline_request_p50 = float(conf['baseline_request_p50']) # avarage
        self.baseline_request_p999 = float(conf['baseline_request_p999']) # percentil 99.99
        self.control_request_p50 = float(conf['control_request_avg'])
        self.control_request_p999 = float(conf['control_request_p50'])
        self.gc_st = float(conf['gc_st'])

        self.env = env
        self.id = id
        self.is_gcing = False
        self.heap = simpy.Container(env)  # our trash heap

        self.requests_arrived = 0
        self.processed_requests = 0
        self.gc_exec_time_sum = 0
        self.collects_performed = 0

    def request_arrived(self, request):
        self.requests_arrived += 1
        request.arrived_at(self.env.now)
        if self.heap.level >= self.gc_st:
            self.env.process(self.run_gc_collect())
        yield self.env.process(self.process_request(request))

    def process_request(self, request):
        yield self.env.process(request.run(self.heap, self.get_service_time()))
        yield self.env.process(request.load_balancer.request_succeeded(request))
        self.processed_requests += 1

    def get_service_time(self):

        if self.is_gcing:
            request_p50, request_p999 = self.baseline_request_p50, self.baseline_request_p999
        else:
            request_p50, request_p999 = self.control_request_p50, self.control_request_p999

        service_time = request_p50 + uniform(0, request_p999 - request_p50)

        return service_time

    def run_gc_collect(self):
        self.is_gcing, before = True, self.env.now
        trash = self.heap.level
        yield self.heap.get(trash)
        yield self.env.timeout(self.gc_time_collecting(trash))  # wait the discarding time

        trash = self.heap.level
        if trash > 0:
            yield self.heap.get(trash)
        self.is_gcing, after = False, self.env.now

        self.gc_exec_time_sum += after - before
        self.collects_performed += 1

    def gc_time_collecting(self, trash):
        return ((trash * 10000000000) * 7.317 * (10 ** -8) + 78.34) / 1000


class ServerControl(ServerBaseline):

    def __init__(self, conf, gci_conf, env, id):
        super().__init__(conf, env, id)

        from .gci import GCI
        self.gci = GCI(gci_conf, self.env, self, )

    def request_arrived(self, request):
        yield self.env.process(self.gci.before(request))

    def process_request(self, request):
        before = request.service_time
        yield self.env.process(super().process_request(request))
        after = request.service_time
        self.gci.request_finished(after - before)
