class Kyouen
  def self.all
    ddb = Aws::DynamoDB::Client.new(
      endpoint: 'http://localhost:8000',
      region: 'ap-northeast-1'
    )
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

    resp.items
  end
end
