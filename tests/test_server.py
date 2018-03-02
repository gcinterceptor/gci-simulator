from .context import LoadBalancer, ServerBaseline, ServerControl, Distribution, Reproduction
import unittest
import simpy


class TestServer(unittest.TestCase):
    @classmethod
    def setUp(self):
        self.env = simpy.Environment()
        create_request_rate = 80
        self.lb = LoadBalancer(create_request_rate, self.env)

    def failure_msg(self, expected, received):
        return "Received value: " + str(received) + ", expected value: " + str(expected)

    def assert_equal(self, expected, received):
        self.assertEqual(expected, received, msg=self.failure_msg(expected, received))

    def test_reproduction_and_distribution(self):
        data = []
        for i in range(1000):
            data.append(i)

        distribution = Distribution(data)
        reproduction = Reproduction(data)

        for i in range(1000):
            self.assert_equal(i, reproduction.next_value())
            self.assertTrue(distribution.next_value() in data)

    def test_server_baseline(self):
        id = 0
        service_time_data = [0.001]
        server = ServerBaseline(self.env, id, service_time_data)
        self.lb.add_server(server)

        self.env.run(0.05)
        self.assert_equal(4, server.requests_arrived)
        self.assert_equal(4, server.processed_requests)
        self.assert_equal(4, len(server.requests))
        self.assert_equal(4, self.lb.succeeded_requests)
        self.assert_equal(4, len(self.lb.requests))

        self.env.run(0.2)
        self.assert_equal(16, server.requests_arrived)
        self.assert_equal(16, server.processed_requests)
        self.assert_equal(16, len(server.requests))
        self.assert_equal(16, self.lb.succeeded_requests)
        self.assert_equal(16, len(self.lb.requests))

    def test_server_control(self):
        id = 0
        service_time_data = [0.001]
        processed_request_data = [3]
        shedded_request_data = [1]
        server = ServerControl(self.env, id, service_time_data, processed_request_data, shedded_request_data)
        self.lb.add_server(server)

        self.env.run(0.05)
        self.assert_equal(3, server.requests_arrived)
        self.assert_equal(3, server.processed_requests)
        self.assert_equal(3, len(server.requests))
        self.assert_equal(4, self.lb.created_requests)
        self.assert_equal(3, self.lb.succeeded_requests)
        self.assert_equal(1, self.lb.shedded_requests)
        self.assert_equal(1, self.lb.lost_requests)
        self.assert_equal(4, len(self.lb.requests))

        self.env.run(0.2)
        self.assert_equal(12, server.requests_arrived)
        self.assert_equal(12, server.processed_requests)
        self.assert_equal(12, len(server.requests))
        self.assert_equal(16, self.lb.created_requests)
        self.assert_equal(12, self.lb.succeeded_requests)
        self.assert_equal(4, self.lb.shedded_requests)
        self.assert_equal(4, self.lb.lost_requests)
        self.assert_equal(16, len(self.lb.requests))


if __name__ == '__main__':
    unittest.main()
