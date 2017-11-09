import unittest
import simpy
from .. modules.clients import Clients, Request

class RequestTest(unittest.TestCase):

    @classmethod
    def setUp(self):
        self.env = simpy.Environment()
        self.heap = simpy.Container(self.env, 100, init=0)  # our trash heap
        self.SIM_DURATION = 10
        self.env.run(until=self.SIM_DURATION)

    def test_run(self):
        #self.env.process(self.__run())
        pass

    def test_set(self):
        self.__sets()

    def __run(self):
        created_at = self.env.now
        request = Request(created_at, 0.035, 0.02, None) # It just keeps who is his owner, but don't do anything with it. That why it can be None.
        self.assertTrue(created_at == request.created_at, "Expected value: " + str(created_at) + ", received value: " + str(request.created_at))

        self.assertTrue(False)  # Deveria quebrar o teste

        time_before = self.env.now
        yield self.env.process(request.run(self.env, self.heap)) # Esse yield está atrapalhando os testes

        # check if the running time of the request is correct
        self.assertTrue(self.env.now - time_before == 0.035, "Expected value: " + str(0.035) + ", received value: " + str(self.env.now - time_before))
        # check if the cost of memory for this case is correct (those values will probably change in future)
        self.assertTrue(self.heap.level == 0.02, "Expected value: " + str(0.02) + ", received value: " + str(self.heap.level))

    def __sets(self):
        created_at = self.env.now
        request = Request(created_at, 0.035, 0.02, None)  # It just keeps who is his owner, but don't do anything with it. That why it can be None.
        # check if the created_at is set correctly
        self.assertTrue(created_at == request.created_at, "Expected value: " + str(created_at) + ", received value: " + str(request.created_at))

        sent_at = self.env.now
        request.sent_at(self.env.now)
        # check if the sent_at is set correctly
        self.assertTrue(sent_at == request._sent, "Expected value: " + str(sent_at) + ", received value: " + str(request._sent))

        yield self.env.timeout(5) # Esse yield está atrapalhando os testes
        self.assertTrue(False) # Deveria quebrar o teste

        done_at = self.env.now
        request.done_at(self.env.now)
        # check if the done_at is set correctly
        self.assertTrue(done_at + 1 == request._done, "Expected value: " + str(done_at) + ", received value: " + str(request._done))

class ClientsTest(unittest.TestCase):

    @classmethod
    def setUp(self):
        #write something...
        pass

    def test_algumaCoisa(self):
        #tests devem começar com 'test'
        pass

if __name__ == '__main__':
    unittest.main()