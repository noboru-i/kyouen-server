FROM ruby:2.3

RUN mkdir /app

WORKDIR /tmp
COPY Gemfile Gemfile
COPY Gemfile.lock Gemfile.lock
RUN bundle install

WORKDIR /app

CMD ["rails", "server", "-b", "0.0.0.0"]
