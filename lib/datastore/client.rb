# frozen_string_literal: true
require 'googleauth'
require 'google/apis/datastore_v1'

module Datastore
  PROJECT = 'my-android-server'
  # https://github.com/google/google-api-ruby-client/blob/master/generated/google/apis/datastore_v1/classes.rb
  ServiceAccountCredentials = Google::Auth::ServiceAccountCredentials
  GCPDatastore = Google::Apis::DatastoreV1
  RunQueryRequest = GCPDatastore::RunQueryRequest
  GqlQuery = GCPDatastore::GqlQuery
  GqlQueryParameter = GCPDatastore::GqlQueryParameter
  Value = GCPDatastore::Value
  CommitRequest = GCPDatastore::CommitRequest
  Mutation = GCPDatastore::Mutation
  Entity = GCPDatastore::Entity
  Key = GCPDatastore::Key
  PathElement = GCPDatastore::PathElement

  class Client
    def initialize
      @datastore = GCPDatastore::DatastoreService.new

      dummy_file = DummyFile.new(ENV['GCP_KEY']) if ENV['GCP_KEY'].present?
      @datastore.authorization = ServiceAccountCredentials.make_creds(
        json_key_io: dummy_file || File.open('cert/my-android-server-91c8d931db89.json'),
        scope: [
          GCPDatastore::AUTH_CLOUD_PLATFORM,
          GCPDatastore::AUTH_DATASTORE
        ]
      )
    end

    def query(query, parameters)
      positional_bindings = parameters.map(&:generate_query_parameter)
      query = GqlQuery.new(
        query_string: query,
        positional_bindings: positional_bindings
      )
      request = RunQueryRequest.new(gql_query: query)
      result = @datastore.run_project_query(PROJECT, request)
      result.batch.entity_results
    end

    def insert # TODO need entity parameter
      transaction_id = @datastore.begin_project_transaction(PROJECT).transaction
      path = PathElement.new(kind: 'User', name: 'KEY1')
      key = Key.new(path: [path])
      properties = {
        'userId': Value.new(string_value: '1')
      }
      print 'properties = ' + properties.to_s
      entity = Entity.new(key: key, properties: properties)
      mutations = [
        Mutation.new(insert: entity)
      ]
      request = CommitRequest.new(transaction: transaction_id, mutations: mutations)
      result = @datastore.commit_project(PROJECT, request)
      print result.to_json.to_s
    end
  end

  class Parameter
    attr_accessor :value

    def initialize(value)
      @value = value
    end

    def generate_query_parameter
      case @value
      when Integer
        GqlQueryParameter.new(value: Value.new(integer_value: @value))
      when String
        GqlQueryParameter.new(value: Value.new(string_value: @value))
      end
    end
  end

  class DummyFile
    attr_accessor :read
    def initialize(read)
      @read = read
    end
  end
end
