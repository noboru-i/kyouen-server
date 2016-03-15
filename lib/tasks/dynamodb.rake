namespace :dynamodb do
  desc 'Create db'
  task create_db: :environment do
    require 'aws-sdk-core'

    ddb = Aws::DynamoDB::Client.new(
      endpoint: 'http://localhost:8000',
      region: 'ap-northeast-1'
    )

    ddb.delete_table({
      table_name: "Kyouen",
    })
    options = {
      attribute_definitions: [
        {
          attribute_name: "id_prefix",
          attribute_type: "N"
        },
        {
          attribute_name: "id",
          attribute_type: "N"
        }
      ],
      table_name: "Kyouen",
      key_schema: [
        {
          attribute_name: "id_prefix",
          key_type: "HASH"
        },
        {
          attribute_name: "id",
          key_type: "RANGE"
        }
      ],
      provisioned_throughput: {
        read_capacity_units: 1,
        write_capacity_units: 1
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
    stages = [
      "1,6,000000010000001100001100000000001000,noboru",
      "2,6,000000000000000100010010001100000000,noboru",
      "3,6,000000001000010000000100010010001000,noboru",
      "4,6,001000001000000010010000010100000000,noboru",
      "5,6,000000001011010000000010001000000010,noboru",
      "6,6,000100000000101011010000000000000000,noboru",
      "7,6,000000001010000000010010000000001010,noboru",
      "8,6,001000000001010000010010000001000000,noboru",
      "9,6,000000001000010000000010000100001000,noboru",
      "10,6,000100000010010000000100000010010000,noboru"
    ]
    stages.each do |stage|
      s = stage.split(',')
      ddb.put_item(
        table_name: 'Kyouen',
        item: {
          id_prefix: s[0].to_i / 100,
          id: s[0].to_i,
          size: s[1].to_i,
          stage: s[2],
          creator: s[3],
          created_at: Time.now.to_i
        }
      )
    end

    start = 1
    resp = ddb.query({
      table_name: "Kyouen",
      select: "ALL_ATTRIBUTES",
      limit: 100,
      key_condition_expression: "id_prefix = :PREFIX AND id BETWEEN :FROM AND :TO",
      expression_attribute_values: {
        ':PREFIX' => start / 100,
        ':FROM' => start,
        ':TO' => start + 9
      },
    })
    puts resp.items
    resp.items.each do |item|
      puts item['id']
    end
  end
end
