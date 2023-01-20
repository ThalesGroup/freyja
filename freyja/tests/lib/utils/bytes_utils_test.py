from freyja.lib.utils.bytes_utils import convert_size


def test_convert_size():
    actual = "4194304 KiB"
    expected = 4.0  # GB
    kib = int(actual.split(" ")[0])

    res = convert_size(kib * 1024)

    assert res
    assert type(res) == str
    assert float(res.split(" ")[0]) == expected
