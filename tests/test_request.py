import sys            # These two first lines, fixes
sys.path.append("..") # the problem of imports from modeles
import unittest, simpy, configparser, os
from modules import Request

class TestRequest(unittest.TestCase):

    @classmethod
    def setUp(self):
        self.config = configparser.ConfigParser()
        self.config.read('../configfiles/request.ini')
        self.env = simpy.Environment()
        self.heap = simpy.Container(self.env, 100, init=0)  # our trash heap
        self.SIM_DURATION = 10

    def test(self):
        created_at = self.env.now
        requests_conf = self.config['request service_time-0.0035 memory-0.02']
        request = Request(created_at, None, requests_conf) # It just keeps who is his owner, but don't do anything with it. That why it can be None.

        sent_at = self.env.now
        request.sent_at(self.env.now)
        self.env.process(request.run(self.env, self.heap))
        self.env_run()

        expected_value = created_at
        # check if the created_at is set correctly
        self.assert_true(expected_value, request.created_at)

        expected_value = request.service_time
        # check if the running time of the request is correct
        self.assert_true(expected_value, request._processed_time)

        expected_value = request.memory
        # check if the cost of memory for this case is correct
        self.assert_true(expected_value, self.heap.level)

        expected_value = sent_at
        # check if the sent_at is set correctly
        self.assert_true(expected_value, request._sent_time)

        expected = 0.000
        self.assert_almost_equal(expected, request.created_at)
        self.assert_almost_equal(expected, request._sent_time)

        expected = 0.0035
        self.assert_almost_equal(expected, request._processed_time)

    def assert_almost_equal(self, expected, received, delta=0.0001):
        msg = "Expected value: " + str(expected) + ", received value: " + str(received)
        self.assertAlmostEqual(expected, received, msg=msg, delta=delta)

    def assert_true(self, expected, received):
        self.assertTrue(expected == received, "Expected value: " + str(expected) + ", received value: " + str(received))

    def env_run(self):
        self.env.run(until=self.SIM_DURATION)


if __name__ == '__main__':
    unittest.main()