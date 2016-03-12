namespace :dynamodb do
  desc 'Create db'
  task create_db: :environment do
    require 'aws-sdk-core'

    ddb = Aws::DynamoDB::Client.new(
        endpoint: 'http://localhost:8000',
        region: 'ap-northeast-1'
    )

    # ddb.delete_table({
    #   table_name: "Kyouen",
    # })
    options = {
      table_name: "Kyouen",
      key_schema: [
          {
              attribute_name: "id",
              key_type: "HASH"
          }
      ],
      attribute_definitions: [
          {
              attribute_name: "id",
              attribute_type: "N"
          }
      ],
      provisioned_throughput: {
          read_capacity_units:  1,
          write_capacity_units:  1
      }
    }
    ddb.create_table(options)
  end

  desc 'Describe db'
  task describe_db: :environment do
    ddb = Aws::DynamoDB::Client.new(
        endpoint: 'http://localhost:8000',
        region: 'ap-northeast-1'
    )
    resp = ddb.describe_table({
      table_name: 'Kyouen'
    })
    p resp
  end

  desc 'Test insert kyouen'
  task test_insert: :environment do
    ddb = Aws::DynamoDB::Client.new(
        endpoint: 'http://localhost:8000',
        region: 'ap-northeast-1'
    )
    ddb.put_item(
      table_name: 'Kyouen',
      item: {
        id:  1,
        size:     6,
        stage:      '000000000000000000000000000000000000',
        creator: 'no name',
        created_at: 123123
      }
    )

    result = ddb.get_item(
      table_name: 'Kyouen',
      key: {
          id: 1
      }
    )
    puts result.item['size']
  end
end
