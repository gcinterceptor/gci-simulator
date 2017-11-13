import sys            # These two first lines, fixes
sys.path.append("..") # the problem of imports from modeles
import unittest, simpy
from modules import Clients, ServerWithGCI, GCI

class TestGCI(unittest.TestCase):
    @classmethod
    def setUp(self):
        self.env = simpy.Environment()
        self.gci = GCI(self.env)
        self.server = ServerWithGCI(self.env, self.gci)
        self.requests = list()

    def test(self):
        pass

    def assert_almost_equal(self, expected, received, delta=0.0001):
        msg = "Expected value: " + str(expected) + ", received value: " + str(received)
        self.assertAlmostEqual(expected, received, msg=msg, delta=delta)

    def env_run(self, sim_duration, max_requests, duration, memory):
        Clients(self.env, self.server, self.requests, max_requests=max_requests, duration=duration, memory=memory)
        self.env.run(until=sim_duration)



if __name__ == '__main__':
    unittest.main()


