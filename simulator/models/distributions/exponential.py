import numpy

class Exponential(object):
    
    def __init__(self, avg, size = 10):
        self.avg = avg
        
        self.size = size
        self.generated_list = list()
    
    def get_next_value(self):
        if(len(self.generated_list) == 0):
            self.generated_list = list(numpy.random.exponential(self.avg, self.size))
        
        return self.generated_list.pop(0)