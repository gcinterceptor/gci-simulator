import unittest
import simpy
from .. modules.clients import Request

class TestRequest(unittest.TestCase):

    @classmethod
    def setUp(self):
        self.env = simpy.Environment()
        self.heap = simpy.Container(self.env, 100, init=0)  # our trash heap
        self.SIM_DURATION = 10

    def test(self):
        created_at = self.env.now
        request = Request(created_at, 0.035, 0.02, None) # It just keeps who is his owner, but don't do anything with it. That why it can be None.

        sent_at = self.env.now
        request.sent_at(self.env.now)
        self.env.process(request.run(self.env, self.heap))
        self.env_run()

        expected_value = created_at
        # check if the created_at is set correctly
        self.assert_true(expected_value, request.created_at)

        expected_value = request.duration
        # check if the running time of the request is correct
        self.assert_true(expected_value, request._processed_time)

        expected_value = request.memory
        # check if the cost of memory for this case is correct
        self.assert_true(expected_value, self.heap.level)

        expected_value = sent_at
        # check if the sent_at is set correctly
        self.assert_true(expected_value, request._sent_time)

    def assert_true(self, expected, received):
        self.assertTrue(expected == received, "Expected value: " + str(expected) + ", received value: " + str(received))

    def env_run(self):
        self.env.run(until=self.SIM_DURATION)


if __name__ == '__main__':
    unittest.main()