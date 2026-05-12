import timeit
from sensor_construct import SensorData

# Тестовые данные
test_data = SensorData.build(dict(
    device_id=12345,
    temperature=23.5,
    humidity=60.0,
    readings_len=10,
    readings=[1,2,3,4,5,6,7,8,9,10]
))

# Бенчмарк
def bench_build():
    return SensorData.build(dict(
        device_id=12345,
        temperature=23.5,
        humidity=60.0,
        readings_len=10,
        readings=[1,2,3,4,5,6,7,8,9,10]
    ))

def bench_parse():
    return SensorData.parse(test_data)

print("Construct build:", timeit.timeit(bench_build, number=1000000), "сек на 1M операций")
print("Construct parse:", timeit.timeit(bench_parse, number=1000000), "сек на 1M операций")
