import simpy
from modules import GC, GCI

class Server(object):

    def __init__(self, env, sleep=0.000001, queue_limite=100, heap_limit=4):
        self.env = env

        self.sleep = sleep
        self.queue = simpy.Store(env, queue_limite) # the queue of requests
        self.remaining_queue = simpy.Store(env, queue_limite)  # the queue of interrupted requests
        self.heap = simpy.Container(env, heap_limit, init=0) # our trash heap

        self.processed_requests = 0
        self.action = env.process(self.run())

        self.gc = GC(self.env, self.heap)
        self.gc_process = self.env.process(self.gc.run(self))

    def run(self):
        try:
            while True:
                if len(self.remaining_queue.items) > 0:
                    request = yield self.remaining_queue.get()  # get a request from store
                    if request.done:
                        yield self.heap.get(request.memory) # remove trash that shouldn't be added...
                    yield self.env.process(self.process_request(request))

                elif len(self.queue.items) > 0:  # check if there is any request to be processed
                    request = yield self.queue.get()  # get a request from store
                    yield self.env.process(self.process_request(request))

                yield self.env.timeout(self.sleep)  # wait for...

        except simpy.Interrupt:
            yield self.remaining_queue.put(request)
            yield self.env.timeout(self.sleep) # wait for...

    def process_request(self, request):
        yield self.env.process(request.run(self.env, self.heap))
        print("At %.3f, SERVER Request processed at server" % self.env.now)
        yield self.env.process(request.client.sucess_request(request))
        self.processed_requests += 1

    def request_arrived(self, request):
        print("At %.3f, SERVER Request stored at Server" % self.env.now)
        yield self.env.process(request.client.successfully_sent(request))
        yield self.queue.put(request)   # put the request at the end of the queue

class ServerWithGCI(Server):

    def __init__(self, env, sleep=0.000001, queue_limite=100, heap_limit=4):
        super().__init__(env, sleep, queue_limite, heap_limit)
        self.gci = GCI(self.env, self)

    def request_arrived(self, request):

        if self.gci.shed_requests:
            print("At %.3f, SERVER Server shedding request" % self.env.now)
            yield self.env.process(request.client.refused_request(request, self.gci.estimated_shed_time()))

        else:
            yield self.queue.put(request)  # put the request at the end of the queue
            self.gci.check()
            yield self.env.timeout(self.sleep)  # wait for...

            print("At %.3f, SERVER Request stored at Server" % self.env.now)
            yield self.env.process(request.client.successfully_sent(request))



