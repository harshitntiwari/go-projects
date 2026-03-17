Requirements:

- Crawler will start with root URL.
- Crawler will recursively:
  - Fetch the data (body + URLs) from this URL
  - Crawl all the URLs received in the step above
- Crawler should not crawl the same URL twice
- Crawler should crawl to a max depth given as input
- Proper error handling if crawler fails to fetch data from a URL (e.g. URL does not exist)

- If you're spawning threads to crawl the URLs concurrently, you should have a fixed size worker pool instead of spawning a thread for every URL
