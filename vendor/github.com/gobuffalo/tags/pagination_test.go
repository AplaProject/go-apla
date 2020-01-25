package tags

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Pagination(t *testing.T) {
	r := require.New(t)

	tag, err := Pagination(&Paginator{
		Page:       2,
		TotalPages: 3,
	}, Options{
		"path": "/foo",
	})
	r.NoError(err)

	r.Equal(template.HTML(`<ul class=" pagination"><li class="page-item"><a class="page-link" href="/foo?page=1">&laquo;</a></li><li class="page-item"><a class="page-link" href="/foo?page=1">1</a></li><li class="page-item active"><a class="page-link" href="/foo?page=2">2</a></li><li class="page-item"><a class="page-link" href="/foo?page=3">3</a></li><li class="page-item"><a class="page-link" href="/foo?page=3">&raquo;</a></li></ul>`), tag.HTML())
}

func Test_Pagination_Page1(t *testing.T) {
	r := require.New(t)

	tag, err := Pagination(&Paginator{
		Page:       1,
		TotalPages: 3,
	}, Options{
		"path": "/foo",
	})
	r.NoError(err)

	r.Equal(template.HTML(`<ul class=" pagination"><li class="page-item disabled"><span class="page-link">&laquo;</span></li><li class="page-item active"><a class="page-link" href="/foo?page=1">1</a></li><li class="page-item"><a class="page-link" href="/foo?page=2">2</a></li><li class="page-item"><a class="page-link" href="/foo?page=3">3</a></li><li class="page-item"><a class="page-link" href="/foo?page=2">&raquo;</a></li></ul>`), tag.HTML())
}

func Test_Pagination_Page3(t *testing.T) {
	r := require.New(t)

	tag, err := Pagination(&Paginator{
		Page:       3,
		TotalPages: 3,
	}, Options{
		"path": "/foo",
	})
	r.NoError(err)

	r.Equal(template.HTML(`<ul class=" pagination"><li class="page-item"><a class="page-link" href="/foo?page=2">&laquo;</a></li><li class="page-item"><a class="page-link" href="/foo?page=1">1</a></li><li class="page-item"><a class="page-link" href="/foo?page=2">2</a></li><li class="page-item active"><a class="page-link" href="/foo?page=3">3</a></li><li class="page-item disabled"><span class="page-link">&raquo;</span></li></ul>`), tag.HTML())
}

func Test_Pagination_LongPageStart(t *testing.T) {
	r := require.New(t)

	tag, err := Pagination(&Paginator{
		Page:       1,
		TotalPages: 29,
	}, Options{
		"path": "/foo",
	})
	r.NoError(err)

	r.Equal(template.HTML(`<ul class=" pagination"><li class="page-item disabled"><span class="page-link">&laquo;</span></li><li class="page-item active"><a class="page-link" href="/foo?page=1">1</a></li><li class="page-item"><a class="page-link" href="/foo?page=2">2</a></li><li class="page-item"><a class="page-link" href="/foo?page=3">3</a></li><li class="page-item"><a class="page-link" href="/foo?page=4">4</a></li><li class="page-item"><a class="page-link" href="/foo?page=5">5</a></li><li class="page-item"><a class="page-link" href="/foo?page=6">6</a></li><li class="page-item"><a class="page-link" href="/foo?page=7">7</a></li><li class="page-item"><a class="page-link" href="/foo?page=8">8</a></li><li class="page-item"><a class="page-link" href="/foo?page=9">9</a></li><li class="page-item disabled"><a>...</a></li><li class="page-item"><a class="page-link" href="/foo?page=29">29</a></li><li class="page-item"><a class="page-link" href="/foo?page=2">&raquo;</a></li></ul>`), tag.HTML())
}

func Test_Pagination_LongPageStartPoint1(t *testing.T) {
	r := require.New(t)

	tag, err := Pagination(&Paginator{
		Page:       6,
		TotalPages: 29,
	}, Options{
		"path": "/foo",
	})
	r.NoError(err)

	r.Equal(template.HTML(`<ul class=" pagination"><li class="page-item"><a class="page-link" href="/foo?page=5">&laquo;</a></li><li class="page-item"><a class="page-link" href="/foo?page=1">1</a></li><li class="page-item"><a class="page-link" href="/foo?page=2">2</a></li><li class="page-item"><a class="page-link" href="/foo?page=3">3</a></li><li class="page-item"><a class="page-link" href="/foo?page=4">4</a></li><li class="page-item"><a class="page-link" href="/foo?page=5">5</a></li><li class="page-item active"><a class="page-link" href="/foo?page=6">6</a></li><li class="page-item"><a class="page-link" href="/foo?page=7">7</a></li><li class="page-item"><a class="page-link" href="/foo?page=8">8</a></li><li class="page-item"><a class="page-link" href="/foo?page=9">9</a></li><li class="page-item disabled"><a>...</a></li><li class="page-item"><a class="page-link" href="/foo?page=29">29</a></li><li class="page-item"><a class="page-link" href="/foo?page=7">&raquo;</a></li></ul>`), tag.HTML())
}

func Test_Pagination_LongPagePoint2(t *testing.T) {
	r := require.New(t)

	tag, err := Pagination(&Paginator{
		Page:       23,
		TotalPages: 29,
	}, Options{
		"path": "/foo",
	})
	r.NoError(err)

	r.Equal(template.HTML(`<ul class=" pagination"><li class="page-item"><a class="page-link" href="/foo?page=22">&laquo;</a></li><li class="page-item"><a class="page-link" href="/foo?page=1">1</a></li><li class="page-item disabled"><a>...</a></li><li class="page-item"><a class="page-link" href="/foo?page=20">20</a></li><li class="page-item"><a class="page-link" href="/foo?page=21">21</a></li><li class="page-item"><a class="page-link" href="/foo?page=22">22</a></li><li class="page-item active"><a class="page-link" href="/foo?page=23">23</a></li><li class="page-item"><a class="page-link" href="/foo?page=24">24</a></li><li class="page-item"><a class="page-link" href="/foo?page=25">25</a></li><li class="page-item"><a class="page-link" href="/foo?page=26">26</a></li><li class="page-item disabled"><a>...</a></li><li class="page-item"><a class="page-link" href="/foo?page=29">29</a></li><li class="page-item"><a class="page-link" href="/foo?page=24">&raquo;</a></li></ul>`), tag.HTML())
}

func Test_Pagination_LongPageEnd(t *testing.T) {
	r := require.New(t)

	tag, err := Pagination(&Paginator{
		Page:       24,
		TotalPages: 29,
	}, Options{
		"path": "/foo",
	})
	r.NoError(err)

	r.Equal(template.HTML(`<ul class=" pagination"><li class="page-item"><a class="page-link" href="/foo?page=23">&laquo;</a></li><li class="page-item"><a class="page-link" href="/foo?page=1">1</a></li><li class="page-item disabled"><a>...</a></li><li class="page-item"><a class="page-link" href="/foo?page=21">21</a></li><li class="page-item"><a class="page-link" href="/foo?page=22">22</a></li><li class="page-item"><a class="page-link" href="/foo?page=23">23</a></li><li class="page-item active"><a class="page-link" href="/foo?page=24">24</a></li><li class="page-item"><a class="page-link" href="/foo?page=25">25</a></li><li class="page-item"><a class="page-link" href="/foo?page=26">26</a></li><li class="page-item"><a class="page-link" href="/foo?page=27">27</a></li><li class="page-item"><a class="page-link" href="/foo?page=28">28</a></li><li class="page-item"><a class="page-link" href="/foo?page=29">29</a></li><li class="page-item"><a class="page-link" href="/foo?page=25">&raquo;</a></li></ul>`), tag.HTML())
}

func Test_Pagination_NextPrevContent(t *testing.T) {
	r := require.New(t)

	tag, err := Pagination(&Paginator{
		Page:       2,
		TotalPages: 3,
	}, Options{
		"path":            "/foo",
		"nextContent":     "Next",
		"previousContent": "Previous",
	})
	r.NoError(err)

	r.Equal(template.HTML(`<ul class=" pagination" nextContent="Next" previousContent="Previous"><li class="page-item"><a class="page-link" href="/foo?page=1">Previous</a></li><li class="page-item"><a class="page-link" href="/foo?page=1">1</a></li><li class="page-item active"><a class="page-link" href="/foo?page=2">2</a></li><li class="page-item"><a class="page-link" href="/foo?page=3">3</a></li><li class="page-item"><a class="page-link" href="/foo?page=3">Next</a></li></ul>`), tag.HTML())
}
