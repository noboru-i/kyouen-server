require 'googleauth'
require 'google/apis/datastore_v1'

module Datastore
  # https://github.com/google/google-api-ruby-client/blob/master/generated/google/apis/datastore_v1/classes.rb
  ServiceAccountCredentials  = Google::Auth::ServiceAccountCredentials
  GCPDatastore  = Google::Apis::DatastoreV1
  RunQueryRequest = GCPDatastore::RunQueryRequest
  GqlQuery = GCPDatastore::GqlQuery
  GqlQueryParameter = GCPDatastore::GqlQueryParameter
  Value = GCPDatastore::Value

  class Client
    def initialize
      @datastore = GCPDatastore::DatastoreService.new

      dummyFile = DummyFile.new(ENV['GCP_KEY']) if ENV['GCP_KEY'].present?
      @datastore.authorization = ServiceAccountCredentials.make_creds(
          json_key_io: dummyFile || File.open('cert/my-android-server-91c8d931db89.json'),
          scope: [
            GCPDatastore::AUTH_CLOUD_PLATFORM,
            GCPDatastore::AUTH_DATASTORE
          ]
      )
    end

    def query(query, parameters)
      positional_bindings = parameters.map{|p|
        p.generate_query_parameter
      }
      query = GqlQuery.new(
          query_string: query,
          positional_bindings: positional_bindings
      )
      request = RunQueryRequest.new(gql_query: query)
      result = @datastore.run_project_query('my-android-server', request)
      return result.batch.entity_results
    end
  end

  class Parameter
    attr_accessor :value

    def initialize(value)
      @value = value
    end

    def generate_query_parameter
      return case @value
      when Fixnum
        GqlQueryParameter.new(value: Value.new(integer_value: @value))
      end
    end
  end

  private class DummyFile
    attr_accessor :read
    def initialize(read)
       @read = read
    end
  end
end
