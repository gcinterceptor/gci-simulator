import unittest
import simpy
from .. modules.clients import Clients, Request

class TestRequest(unittest.TestCase):

    @classmethod
    def setUp(self):
        self.env = simpy.Environment()
        self.heap = simpy.Container(self.env, 100, init=0)  # our trash heap
        self.SIM_DURATION = 10

    def test(self):
        created_at = self.env.now
        request = Request(created_at, 0.035, 0.02, None) # It just keeps who is his owner, but don't do anything with it. That why it can be None.

        # check if the created_at is set correctly
        self.assertTrue(created_at == request.created_at, "Expected value: " + str(created_at) + ", received value: " + str(request.created_at))

        sent_at = self.env.now
        request.sent_at(self.env.now)
        self.env.process(request.run(self.env, self.heap))
        self.env_run()

        expected_value = request.duration
        # check if the running time of the request is correct
        self.assertTrue(request._processed_time == expected_value, "Expected value: " + str(expected_value) + ", received value: " + str(request._processed_time))

        expected_value = request.memory
        # check if the cost of memory for this case is correct (those values will probably change in future)
        self.assertTrue(self.heap.level == expected_value, "Expected value: " + str(expected_value) + ", received value: " + str(self.heap.level))

        expected_value = sent_at
        # check if the sent_at is set correctly
        self.assertTrue(expected_value == request._sent_time, "Expected value: " + str(expected_value) + ", received value: " + str(request._sent_time))

    def env_run(self):
        self.env.run(until=self.SIM_DURATION)



class TestClients(unittest.TestCase):

    @classmethod
    def setUp(self):
        #write something...
        pass

    def test_algumaCoisa(self):
        #tests devem começar com 'test'
        pass

if __name__ == '__main__':
    unittest.main()