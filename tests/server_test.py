import sys            # These two first lines, fixes
sys.path.append("..") # the problem of imports from modeles

import unittest
from modules.server import Server, ServerWithGCI

class ServerTest(unittest.TestCase):

    @classmethod
    def setUp(self):
        pass

    def test_algumaCoisa(self):
        pass

class ServerWithGCITest(unittest.TestCase):

    @classmethod
    def setUp(self):
        pass

    def test_algumaCoisa(self):
        pass


if __name__ == '__main__':
    unittest.main()