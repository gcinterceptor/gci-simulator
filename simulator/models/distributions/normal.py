import numpy

class Normal(Distribution):
    
    def __init__(self, avg, desviation, size = 10):
        self.avg = avg
        self.desviation = desviation
        
        self.size = size
        self.generated_list = list()
    
    def get_next_value(self):
        if(len(self.generated_list) == 0):
            self.generated_list = numpy.random.normal(self.avg, self.desviation, self.size) 
        
        return self.generated_list.pop(0)