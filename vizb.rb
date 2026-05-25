class Vizb < Formula
  desc "Transform Go, Rust, and JS benchmark output into interactive 4D visualizations"
  homepage "https://vizb.goptics.org/"
  url "https://github.com/goptics/vizb/archive/refs/tags/v0.11.0.tar.gz"
  sha256 "d378286151bf4f7612ac4a1017178852d4aff0d19ede7def36fd04b26d3f44ca"
  license "MIT"
  head "https://github.com/goptics/vizb.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X github.com/goptics/vizb/version.Version=v#{version}")
  end

  test do
    (testpath/"bench.txt").write <<~EOS
      BenchmarkFibonacci-8   	    5000	    230456 ns/op
      BenchmarkFibonacci-8   	    5000	    231234 ns/op
    EOS

    system bin/"vizb", "--output", testpath/"result.html", testpath/"bench.txt"
    assert_path_exists testpath/"result.html"
  end
end
