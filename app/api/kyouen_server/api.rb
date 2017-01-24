# frozen_string_literal: true

module SharedParams
  extend Grape::API::Helpers

  params :pagination do
    optional :offset, type: Integer, default: 0
    optional :limit, type: Integer, default: 10, max_value: 100
  end
end

module KyouenServer
  module Entities
    class Stage < Grape::Entity
      expose :stage_no, documentation: { type: :integer, desc: 'Stage number.', required: true }
      expose :size, documentation: { type: :integer, desc: 'Size of stage.', required: true }
      expose :stage, documentation: { type: :string, desc: 'Describe stage state.', required: true }
      expose :creator, documentation: { type: :string, desc: 'Name of who created this.' }
    end
  end

  class API < Grape::API
    version 'v1', using: :header, vendor: 'kyouen'
    format :json

    require_relative '../validations/max_value'
    helpers SharedParams

    resource :users do
      desc 'Login.'
      params do
        requires :token
        requires :token_secret
      end
      post :login do
        twitter_user = TwitterModel.new(params[:token], params[:token_secret]).me
        raise 'twitter login error' if twitter_user.blank?
        user = User.find_by_user_id(twitter_user.id)
        if user.blank?
          user = User.create_new_user(twitter_user.id, twitter_user.screen_name, twitter_user.profile_image_url_https)
        end
        user.generate_api_token
        user
      end
    end

    resource :stages do
      desc 'Return list of stages.', \
           entity: Entities::Stage,
           success: { code: 200, model: Entities::Stage }
      params do
        use :pagination
      end
      get '/' do
        list = KyouenPuzzle.fetch(params[:offset].to_i, params[:limit].to_i)
        present list, with: Entities::Stage
      end
    end

    resource :answers do
      desc 'Send answer of stage.', \
           success: { code: 200, model: Entities::Stage }
      params do
        requires :stage_no, type: Integer
        requires :stage, type: String
      end
      post '/' do
        # check(params[:stage_no], params[:stage])
      end
    end

    add_swagger_documentation \
      info: {
        title: 'API documentation of kyouen app.'
      }
  end
end
