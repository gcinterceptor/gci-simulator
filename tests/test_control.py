from .context import ClientLB, ServerControl, get_config
import unittest
import simpy


class TestControl(unittest.TestCase):

    @classmethod
    def setUp(self):
        gc_conf = get_config('config/gcc.ini', 'gcc sleep_time-0.00001 threshold-0.9 collect_duration-0.25 delay-1')
        req_conf = get_config('config/request.ini', 'request service_time-0.006 memory-0.001555556')
        lb_conf = get_config('config/clientlb.ini', 'clientlb sleep_time-0.00001 create_request_rate-150 max_requests-inf')
        gci_conf = get_config('config/gci.ini', 'gci sleep_time-0.00001 threshold-0.7 check_heap-10 initial_eget-0.9')
        server_conf = get_config('config/server.ini', 'server sleep_time-0.00001')

        self.env = simpy.Environment()
        self.load_balancer = ClientLB(self.env, lb_conf, req_conf)
        self.server = ServerControl(self.env, 1, server_conf, gc_conf, gci_conf)
        self.load_balancer.add_server(self.server)

    def failure_msg(self, expected, received):
        return "Received value: " + str(received) + ", expected value: " + str(expected)

    def assert_equal(self, expected, received):
        self.assertEqual(expected, received, msg=self.failure_msg(expected, received))

    def assert_almost_equal(self, expected, received, delta=0.0001):
        self.assertAlmostEqual(expected, received, msg=self.failure_msg(expected, received), delta=delta)

    def test_simulation_flow(self):
        self.env.run(3)
        expected = 0
        shedded_requests = self.load_balancer.shedded_requests
        self.assert_equal(expected, shedded_requests)
        expected = 450  # 150 succeeded requests by second, after 3 seconds.
        created_requests = self.load_balancer.created_requests
        self.assert_equal(expected, created_requests)
        requests = self.load_balancer.requests
        succeeded_requests = self.load_balancer.succeeded_requests
        self.assert_equal(expected, len(requests))
        self.assert_equal(expected, succeeded_requests)

        self.env.run(4)
        expected = 600
        created_requests = self.load_balancer.created_requests
        self.assert_equal(expected, created_requests)
        created_requests = self.load_balancer.created_requests
        self.assert_equal(expected, created_requests)
        expected = 38
        shedded_requests = self.load_balancer.shedded_requests
        self.assert_equal(expected, shedded_requests)
        expected = 562 # since pass 0.25 gcing, sheeding 38 requests.
        requests = self.load_balancer.requests
        succeeded_requests = self.load_balancer.succeeded_requests
        self.assert_equal(expected, len(requests))
        self.assert_equal(expected, succeeded_requests)

        self.env.run(5)
        expected = 750
        created_requests = self.load_balancer.created_requests
        self.assert_equal(expected, created_requests)
        expected = 38
        shedded_requests = self.load_balancer.shedded_requests
        self.assert_equal(expected, shedded_requests)
        expected = 712
        requests = self.load_balancer.requests
        succeeded_requests = self.load_balancer.succeeded_requests
        self.assert_equal(expected, len(requests))
        self.assert_equal(expected, succeeded_requests)

        self.env.run(7)
        expected = 1050
        created_requests = self.load_balancer.created_requests
        self.assert_equal(expected, created_requests)
        created_requests = self.load_balancer.created_requests
        self.assert_equal(expected, created_requests)
        expected = 76
        shedded_requests = self.load_balancer.shedded_requests
        self.assert_equal(expected, shedded_requests)
        expected = 974
        requests = self.load_balancer.requests
        succeeded_requests = self.load_balancer.succeeded_requests
        self.assert_equal(expected, len(requests))
        self.assert_equal(expected, succeeded_requests)

    def test_flags_variables(self):
        self.env.run(7)

        expected = 974
        processed_requests = self.server.processed_requests
        requests_arrived = self.server.requests_arrived
        self.assert_equal(expected, processed_requests)
        self.assert_equal(expected, requests_arrived)
        expected = 0
        times_interrupted = self.server.times_interrupted
        self.assert_equal(expected, times_interrupted)

        expected = 0
        times_performed = self.server.gc.times_performed
        self.assert_equal(expected, times_performed)
        expected = 2
        collects_performed = self.server.gc.collects_performed
        self.assert_equal(expected, collects_performed)
        expected = 0.5
        gc_exec_time_sum = self.server.gc.gc_exec_time_sum
        self.assert_almost_equal(expected, gc_exec_time_sum)

        expected = 74
        requests_to_process = self.server.gci.requests_to_process
        processed_requests = self.server.gci.processed_requests
        self.assert_equal(expected, requests_to_process)
        self.assert_equal(expected, processed_requests)
        times_performed = self.server.gci.times_performed
        expected = 2
        self.assert_equal(expected, times_performed)

    def test_request_list(self):
        self.env.run(20)
        requests = self.load_balancer.requests

        id_set = set()
        created_time_set = set()
        arrived_time_set = set()
        attended_time_set = set()
        finished_time_set = set()

        for request in requests:
            self.assertTrue(request.done)
            id_set.add(request.id)
            created_time_set.add(request.created_at)
            arrived_time_set.add(request._arrived_time)
            attended_time_set.add(request._attended_time)
            finished_time_set.add(request._finished_time)

        expected = len(requests)
        self.assert_equal(expected, len(id_set))
        self.assert_equal(expected, len(created_time_set))
        self.assert_equal(expected, len(arrived_time_set))
        self.assert_equal(expected, len(attended_time_set))
        self.assert_equal(expected, len(finished_time_set))


if __name__ == '__main__':
    unittest.main()


