import unittest

def hello_world():
    return "Hello, World!"

class TestHelloType(unittest.TestCase):
    def test_hello_type(self):
        result = hello_world()
        self.assertIsInstance(result, str)

if __name__ == '__main__':
    unittest.main()
---
